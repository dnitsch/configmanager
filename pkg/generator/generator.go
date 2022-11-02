package generator

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"regexp"
	"strings"
	"sync"

	"github.com/dnitsch/configmanager/pkg/log"
)

const (
	// tokenSeparator used for identifying the end of a prefix and beginning of token
	// see notes about special consideration for AZKVSECRET tokens
	tokenSeparator = "#"
	// keySeparator used for accessing nested objects within the retrieved map
	keySeparator = "|"
	// AWS SecretsManager prefix
	SecretMgrPrefix = "AWSSECRETS"
	// AWS Parameter Store prefix
	ParamStorePrefix = "AWSPARAMSTR"
	// Azure Key Vault Secrets prefix
	AzKeyVaultSecretsPrefix = "AZKVSECRET"
)

var (
	// default varPrefix used by the replacer function
	// any token msut beging with one of these else
	// it will be skipped as not a replaceable token
	VarPrefix = map[string]bool{SecretMgrPrefix: true, ParamStorePrefix: true, AzKeyVaultSecretsPrefix: true}
)

// Generatoriface describes the exported methods
// on the GenVars struct.
type Generatoriface interface {
	Generate(tokens []string) (ParsedMap, error)
	ConvertToExportVar() []string
	FlushToFile(w io.Writer) error
	StrToFile(w io.Writer, str string) error
}

// GenVarsiface stores strategy and GenVars implementation behaviour
type GenVarsiface interface {
	Generatoriface
	Config() *GenVarsConfig
	ConfigOutputPath() string
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
	// rawMap is the internal object that holds the values of original token => retrieved value - decrypted in plain text
	rawMap ParsedMap
}

// setValue implements GenVarsiface
func (*GenVars) setValue(s string) {
	panic("unimplemented")
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
	m := ParsedMap{}
	defaultConf := GenVarsConfig{
		tokenSeparator: tokenSeparator,
		keySeparator:   keySeparator,
	}
	return &GenVars{
		rawMap: m,
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

// ConfigOutputPath returns the output path set on GenVars create
// withconfig or default value
func (c *GenVars) ConfigOutputPath() string {
	return c.config.outpath
}

// GenVarsConfig defines the input config object to be passed
type GenVarsConfig struct {
	outpath        string
	tokenSeparator string
	keySeparator   string
}

// NewConfig
func NewConfig() *GenVarsConfig {
	return &GenVarsConfig{
		tokenSeparator: tokenSeparator,
		keySeparator:   keySeparator,
	}
}

// WithOutputPath
func (c *GenVarsConfig) WithOutputPath(out string) *GenVarsConfig {
	c.outpath = out
	return c
}

// WithTokenSeparator adds a custom token separator
// token is the actual value of the parameter/secret in the
// provider store
func (c *GenVarsConfig) WithTokenSeparator(tokenSeparator string) *GenVarsConfig {
	c.tokenSeparator = tokenSeparator
	return c
}

// WithKeySeparator adds a custom key separotor
func (c *GenVarsConfig) WithKeySeparator(keySeparator string) *GenVarsConfig {
	c.keySeparator = keySeparator
	return c
}

// Generate generates a k/v map of the tokens with their corresponding secret/paramstore values
// the standard pattern of a token should follow a path like
func (c *GenVars) Generate(tokens []string) (ParsedMap, error) {

	rawTokenPrefixMap := map[string]string{}
	for _, token := range tokens {
		prefix := strings.Split(token, c.config.tokenSeparator)[0]
		if found := VarPrefix[prefix]; found {
			rawTokenPrefixMap[token] = prefix
		}
	}
	pm, err := c.generate(rawTokenPrefixMap)
	if err != nil {
		return nil, err
	}
	c.rawMap = pm
	return pm, nil
}

type chanResp struct {
	value string
	key   string
	err   error
}

// generate checks if any tokens found
// initiates groutines with fixed size channel map
// to capture responses and errors
// generates ParsedMap which includes
func (c *GenVars) generate(rawMap map[string]string) (ParsedMap, error) {
	outMap := ParsedMap{}
	if len(rawMap) < 1 {
		log.Debug("no replaceable tokens found in input strings")
		return map[string]any{}, nil
	}

	var errors []error
	// build an exact size channel
	var wg sync.WaitGroup
	initChanLen := len(rawMap)
	outCh := make(chan chanResp, initChanLen)

	wg.Add(initChanLen)
	for token, prefix := range rawMap {
		go func(a, p string) {
			defer wg.Done()
			rs := newRetrieveStrategy(NewDefatultStrategy(), c.config)
			outCh <- rs.retrieveSpecificCh(c.ctx, p, a)
		}(token, prefix)
	}

	go func() {
		wg.Wait()
		close(outCh)
	}()

	for cro := range outCh {
		log.Debugf("cro: %+v", cro)
		cr := cro
		if cr.err != nil {
			log.Debugf("cr.err %v, for token: %s", cr.err, cr.key)
			errors = append(errors, cr.err)
		}
		outMap[cr.key] = c.keySeparatorLookup(cr.key, cr.value)
		log.Debugf("outMap iter: %+v", outMap)

	}

	if len(errors) > 0 {
		// crude ...
		log.Debugf("found: %d errors", len(errors))
		// return c.rawMap, fmt.Errorf("%+v", errors)
	}
	log.Debugf("complete outMap: %+v", outMap)
	return outMap, nil
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

	pm := ParsedMap{}

	if ok := isParsed(val, &pm); ok {
		log.Debugf("attempting to find by key: %v in value: %v", kl, val)
		if foundVal, ok := pm[kl[1]]; ok {
			log.Debugf("found by key: %v, in value: %v, of: %v", kl[1], val, foundVal)
			return fmt.Sprintf("%v", foundVal)
		}
	}
	// returns the input value string as is
	return val
}

// ConvertToExportVar assigns the k/v out
// as unix style export key=val pairs separated by `\n`
func (c *GenVars) ConvertToExportVar() []string {
	for k, v := range c.rawMap {
		rawKeyToken := strings.Split(k, "/") // assumes a path like token was used
		topLevelKey := rawKeyToken[len(rawKeyToken)-1]
		trm := &ParsedMap{}
		isOk := isParsed(v, trm)
		if isOk {
			// if is a map
			// try look up on key if separator defined
			normMap := c.envVarNormalize(*trm)
			c.exportVars(normMap)
		} else {
			c.exportVars(ParsedMap{topLevelKey: v})
		}
	}
	return c.outString
}

// envVarNormalize
func (c *GenVars) envVarNormalize(pmap ParsedMap) ParsedMap {
	normalizedMap := ParsedMap{}
	for k, v := range pmap {
		normalizedMap[c.normalizeKey(k)] = v
	}
	return normalizedMap
}

func (c *GenVars) exportVars(exportMap ParsedMap) {

	for k, v := range exportMap {
		// NOTE: \n line ending is not totally cross platform
		_type := fmt.Sprintf("%T", v)
		switch _type {
		case "string":
			c.outString = append(c.outString, fmt.Sprintf("export %s='%s'", c.normalizeKey(k), v))
		default:
			c.outString = append(c.outString, fmt.Sprintf("export %s=%v", c.normalizeKey(k), v))
		}
	}
}

// normalizeKeys returns env var compatible key
func (c *GenVars) normalizeKey(k string) string {
	// TODO: include a more complete regex of vaues to replace
	r := regexp.MustCompile(`[\s\@\!]`).ReplaceAll([]byte(k), []byte(""))
	r = regexp.MustCompile(`[\-]`).ReplaceAll(r, []byte("_"))
	// Double underscore replace key separator
	r = regexp.MustCompile(`[`+c.config.keySeparator+`]`).ReplaceAll(r, []byte("__"))
	return strings.ToUpper(string(r))
}

// FlushToFile saves contents to file provided
// in the config input into the generator
// default location is ./app.env
func (c *GenVars) FlushToFile(w io.Writer) error {
	return c.flushToFile(w, listToString(c.outString))
}

// StrToFile
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
