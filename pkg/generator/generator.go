package generator

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"
	"sync"

	"github.com/dnitsch/configmanager/pkg/log"
	"github.com/spyzhov/ajson"
)

type ImplementationPrefix string

const (
	// AWS SecretsManager prefix
	SecretMgrPrefix ImplementationPrefix = "AWSSECRETS"
	// AWS Parameter Store prefix
	ParamStorePrefix ImplementationPrefix = "AWSPARAMSTR"
	// Azure Key Vault Secrets prefix
	AzKeyVaultSecretsPrefix ImplementationPrefix = "AZKVSECRET"
	// Azure Key Vault Secrets prefix
	AzTableStorePrefix ImplementationPrefix = "AZTABLESTORE"
	// Azure App Config prefix
	AzAppConfigPrefix ImplementationPrefix = "AZAPPCONF"
	// Hashicorp Vault prefix
	HashicorpVaultPrefix ImplementationPrefix = "VAULT"
	// GcpSecrets
	GcpSecretsPrefix ImplementationPrefix = "GCPSECRETS"
)

const (
	// tokenSeparator used for identifying the end of a prefix and beginning of token
	// see notes about special consideration for AZKVSECRET tokens
	tokenSeparator = "#"
	// keySeparator used for accessing nested objects within the retrieved map
	keySeparator = "|"
)

var (
	// default varPrefix used by the replacer function
	// any token must beging with one of these else
	// it will be skipped as not a replaceable token
	VarPrefix = map[ImplementationPrefix]bool{
		SecretMgrPrefix: true, ParamStorePrefix: true, AzKeyVaultSecretsPrefix: true,
		GcpSecretsPrefix: true, HashicorpVaultPrefix: true, AzTableStorePrefix: true,
		AzAppConfigPrefix: true,
	}
)

// Generatoriface describes the exported methods
// on the GenVars struct.
type Generatoriface interface {
	Generate(tokens []string) (ParsedMap, error)
	ConvertToExportVar() []string
	FlushToFile(w io.Writer, outString []string) error
	StrToFile(w io.Writer, str string) error
}

// GenVarsiface stores strategy and GenVars implementation behaviour
type GenVarsiface interface {
	Generatoriface
	Config() *GenVarsConfig
}

type muRawMap struct {
	mu       sync.RWMutex
	tokenMap ParsedMap
}

// GenVars is the main struct holding the
// strategy patterns iface
// any initialised config if overridded with withers
// as well as the final outString and the initial rawMap
// which wil be passed in a loop into a goroutine to perform the
// relevant strategy network calls to the config store implementations
type GenVars struct {
	Generatoriface
	ctx       context.Context
	config    GenVarsConfig
	outString []string
	// rawMap is the internal object that holds the values
	// of original token => retrieved value - decrypted in plain text
	// with a mutex RW locker
	rawMap muRawMap //ParsedMap
}

// ParsedMap is the internal working object definition and
// the return type if results are not flushed to file
type ParsedMap map[string]any

// NewGenerator returns a new instance of Generator
// with a default strategy pattern wil be overwritten
// during the first run of a found tokens map
func NewGenerator() *GenVars {
	// defaultStrategy := NewDefatultStrategy()
	return newGenVars()
}

func newGenVars() *GenVars {
	m := make(ParsedMap)
	defaultConf := GenVarsConfig{
		tokenSeparator: tokenSeparator,
		keySeparator:   keySeparator,
	}
	return &GenVars{
		rawMap: muRawMap{tokenMap: m},
		ctx:    context.TODO(),
		// return using default config
		config: defaultConf,
	}
}

// WithConfig uses custom config
func (c *GenVars) WithConfig(cfg *GenVarsConfig) *GenVars {
	// backwards compatibility
	if cfg != nil {
		c.config = *cfg
	}
	return c
}

// WithContext uses caller passed context
func (c *GenVars) WithContext(ctx context.Context) *GenVars {
	c.ctx = ctx
	return c
}

// Config gets Config on the GenVars
func (c *GenVars) Config() *GenVarsConfig {
	return &c.config
}

func (c *GenVars) RawMap() ParsedMap {
	c.rawMap.mu.RLock()
	defer c.rawMap.mu.RUnlock()
	// make a copy of the map
	m := make(ParsedMap)
	for k, v := range c.rawMap.tokenMap {
		m[k] = v
	}
	return m
}

func (c *GenVars) AddRawMap(key, val string) {
	c.rawMap.mu.Lock()
	defer c.rawMap.mu.Unlock()
	c.rawMap.tokenMap[key] = c.keySeparatorLookup(key, val)
}

// Generate generates a k/v map of the tokens with their corresponding secret/paramstore values
// the standard pattern of a token should follow a path like
func (c *GenVars) Generate(tokens []string) (ParsedMap, error) {

	rawTokenPrefixMap := make(map[string]string)
	for _, token := range tokens {
		prefix := strings.Split(token, c.config.tokenSeparator)[0]
		if found := VarPrefix[ImplementationPrefix(prefix)]; found {
			rawTokenPrefixMap[token] = prefix
		}
	}
	rs := newRetrieveStrategy(NewDefatultStrategy(), c.config)
	// pass in default initialised retrieveStrategy
	if err := c.generate(rawTokenPrefixMap, rs); err != nil {
		return nil, err
	}
	return c.RawMap(), nil
}

type chanResp struct {
	value string
	key   string
	err   error
}

type retrieveIface interface {
	RetrieveByToken(ctx context.Context, impl genVarsStrategy, prefix ImplementationPrefix, in string) chanResp
	SelectImplementation(ctx context.Context, prefix ImplementationPrefix, in string, config GenVarsConfig) (genVarsStrategy, error)
}

// generate checks if any tokens found
// initiates groutines with fixed size channel map
// to capture responses and errors
// generates ParsedMap which includes
func (c *GenVars) generate(rawMap map[string]string, rs retrieveIface) error {
	if len(rawMap) < 1 {
		log.Debug("no replaceable tokens found in input strings")
		return nil
	}

	var errors []error
	// build an exact size channel
	var wg sync.WaitGroup
	initChanLen := len(rawMap)
	outCh := make(chan chanResp, initChanLen)

	wg.Add(initChanLen)
	// TODO: initialise the singleton serviceContainer
	// pass into each goroutine
	for token, prefix := range rawMap {
		// take value from config allocation on a per iteration basis
		conf := c.Config()
		go func(prfx ImplementationPrefix, tkn string, conf GenVarsConfig) {
			defer wg.Done()
			strategy, err := rs.SelectImplementation(c.ctx, prfx, tkn, conf)
			if err != nil {
				outCh <- chanResp{err: err}
				return
			}
			outCh <- rs.RetrieveByToken(c.ctx, strategy, prfx, tkn)
		}(ImplementationPrefix(prefix), token, *conf)
	}

	go func() {
		wg.Wait()
		close(outCh)
	}()

	for cro := range outCh {
		cr := cro
		log.Debugf("cro: %+v", cr)
		if cr.err != nil {
			log.Debugf("cr.err %v, for token: %s", cr.err, cr.key)
			errors = append(errors, cr.err)
			// Skip adding not found key to the RawMap
			continue
		}
		c.AddRawMap(cr.key, cr.value)
	}

	if len(errors) > 0 {
		// crude ...
		log.Debugf("found: %d errors", len(errors))
		// return outMap, fmt.Errorf("%v", errors)
	}
	return nil
}

// isParsed will try to parse the return found string into
// map[string]string
// If found it will convert that to a map with all keys uppercased
// and any characters
func isParsed(res any, trm *ParsedMap) bool {
	str := fmt.Sprint(res)
	if err := json.Unmarshal([]byte(str), &trm); err != nil {
		log.Info("unable to parse into a k/v map returning a string instead")
		return false
	}
	log.Info("parsed into a k/v map")
	return true
}

// keySeparatorLookup checks if the key contains
// keySeparator character
// If it does contain one then it tries to parse
func (c *GenVars) keySeparatorLookup(key, val string) string {
	// key has separator
	kl := strings.Split(key, c.config.keySeparator)
	log.Debugf("key list: %v", kl)
	if len(kl) < 2 {
		log.Infof("no keyseparator found")
		return val
	}

	keys, err := ajson.JSONPath([]byte(val), fmt.Sprintf("$..%s", kl[1]))
	if err != nil {
		log.Debugf("unable to parse as json object %v", err.Error())
		return val
	}

	if len(keys) == 1 {
		v := keys[0]
		if v.Type() == ajson.String {
			str, err := strconv.Unquote(fmt.Sprintf("%v", v))
			if err != nil {
				log.Debugf("unable to unquote value: %v returning as is", v)
				return fmt.Sprintf("%v", v)
			}
			return str
		}

		return fmt.Sprintf("%v", v)
	}

	log.Infof("no value found in json using path expression")
	return ""
}

// ConvertToExportVar assigns the k/v out
// as unix style export key=val pairs separated by `\n`
func (c *GenVars) ConvertToExportVar() []string {
	for k, v := range c.RawMap() {
		rawKeyToken := strings.Split(k, "/") // assumes a path like token was used
		topLevelKey := rawKeyToken[len(rawKeyToken)-1]
		trm := make(ParsedMap)
		if parsedOk := isParsed(v, &trm); parsedOk {
			// if is a map
			// try look up on key if separator defined
			normMap := c.envVarNormalize(trm)
			c.exportVars(normMap)
			continue
		}
		c.exportVars(ParsedMap{topLevelKey: v})
	}
	return c.outString
}

// envVarNormalize
func (c *GenVars) envVarNormalize(pmap ParsedMap) ParsedMap {
	normalizedMap := make(ParsedMap)
	for k, v := range pmap {
		normalizedMap[c.normalizeKey(k)] = v
	}
	return normalizedMap
}

func (c *GenVars) exportVars(exportMap ParsedMap) {

	for k, v := range exportMap {
		// NOTE: \n line ending is not totally cross platform
		t := fmt.Sprintf("%T", v)
		switch t {
		case "string":
			c.outString = append(c.outString, fmt.Sprintf("export %s='%s'", c.normalizeKey(k), v))
		default:
			c.outString = append(c.outString, fmt.Sprintf("export %s=%v", c.normalizeKey(k), v))
		}
	}
}

// normalizeKeys returns env var compatible key
func (c *GenVars) normalizeKey(k string) string {
	// the order of replacer pairs matters less
	// as the Replace builds a node tree without overlapping matches
	replacer := strings.NewReplacer([]string{" ", "", "@", "", "!", "", "-", "_", c.config.keySeparator, "__"}...)
	return strings.ToUpper(replacer.Replace(k))
}

// FlushToFile saves contents to file provided
// in the config input into the generator
// default location is ./app.env
func (c *GenVars) FlushToFile(w io.Writer, out []string) error {
	return c.flushToFile(w, listToString(c.outString))
}

// StrToFile writes a provided string to the writer
func (c *GenVars) StrToFile(w io.Writer, str string) error {
	return c.flushToFile(w, str)
}

func (c *GenVars) flushToFile(f io.Writer, str string) error {

	_, e := f.Write([]byte(str))

	if e != nil {
		return e
	}

	return nil
}

func listToString(strList []string) string {
	return strings.Join(strList, "\n")
}
