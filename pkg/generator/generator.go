package generator

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/dnitsch/configmanager/pkg/log"
)

const (
	tokenSeparator = "#"
	// keySeparator used for accessing nested
	// objects within the retrieved map
	keySeparator = "|"
	//
	// secretMgrPrefix  = "AWSSECRETS"
	// paramStorePrefix = "AWSPARAMSTR"
	SecretMgrPrefix  = "AWSSECRETS"
	ParamStorePrefix = "AWSPARAMSTR"
	AzKeyVaultPrefix = "AZKEYVAULT"
)

var (
	// default varPrefix
	VarPrefix = map[string]bool{SecretMgrPrefix: true, ParamStorePrefix: true}
)

type Generatoriface interface {
	Generate(tokens []string) (ParsedMap, error)
	ConvertToExportVar()
	FlushToFile() (string, error)
}

type GenVars struct {
	Generatoriface
	implementation genVarsStrategy
	token          string
	ctx            context.Context
	config         GenVarsConfig
	outString      []string
	// rawMap is the internal object that holds the values of original token => retrieved value - decrypted in plain text
	rawMap ParsedMap
}

// ParsedMap is the internal working object definition and
// the return type if results are not flushed to file
type ParsedMap map[string]any

// NewGenerator returns a new instance
func NewGenerator() *GenVars {
	defaultStrategy := NewDefatultStrategy()
	return newGenVars(defaultStrategy)
}

func newGenVars(e genVarsStrategy) *GenVars {
	m := ParsedMap{}
	defaultConf := GenVarsConfig{
		tokenSeparator: tokenSeparator,
		keySeparator:   keySeparator,
	}
	return &GenVars{
		implementation: e,
		rawMap:         m,
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

// GenVarsConfig defines the input config object to be passed
type GenVarsConfig struct {
	outpath        string
	tokenSeparator string
	keySeparator   string
	// varPrefix      VarPrefix
}

// NewConfig
func NewConfig() *GenVarsConfig {
	return &GenVarsConfig{
		tokenSeparator: tokenSeparator,
		keySeparator:   keySeparator,
		// varPrefix:      varPrefix,
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

	for _, token := range tokens {
		prefix := strings.Split(token, c.config.tokenSeparator)[0]
		if found := VarPrefix[prefix]; found {
			// TODO: allow for more customization here
			rawString, err := c.retrieveSpecific(prefix, token)
			if err != nil {
				return nil, err
			}
			// check if token includes keySeparator
			c.rawMap[token] = c.keySeparatorLookup(token, rawString)
		}
	}
	return c.rawMap, nil
}

func (c *GenVars) retrieveSpecific(prefix, in string) (string, error) {
	switch prefix {
	case SecretMgrPrefix:
		// default strategy paramstore
		scrtMgr, err := NewSecretsMgr(c.ctx)
		if err != nil {
			return "", err
		}
		c.setImplementation(scrtMgr)
		c.setToken(in)
		return c.getTokenValue()
	case ParamStorePrefix:
		paramStr, err := NewParamStore(c.ctx)
		if err != nil {
			return "", err
		}
		c.setImplementation(paramStr)
		c.setToken(in)
		return c.getTokenValue()
	case AzKeyVaultPrefix:
		azKv, err := NewKvStoreWithToken(c.ctx, in, c.config.tokenSeparator, c.config.keySeparator)
		if err != nil {
			return "", err
		}
		// Need to swap this around for AzKV as the
		// client is initialised via vaultURI
		//
		c.setImplementation(azKv)
		c.setToken(in)
		return c.getTokenValue()
	default:
		return "", fmt.Errorf("implementation not found for input string: %s", in)
	}
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
	if len(kl) < 2 {
		return val
	}

	pmptr := &ParsedMap{}

	if ok := isParsed(val, pmptr); ok {
		pm := *pmptr
		if foundVal, ok := pm[kl[1]]; ok {
			return fmt.Sprintf("%v", foundVal)
		}
	}
	// returns the input value string as is
	return val
}

// ConvertToExportVar assigns the k/v out
// as unix style export key=val pairs seperated by `\n`
func (c *GenVars) ConvertToExportVar() {
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

func (c *GenVars) stripPrefix(in, prefix string) string {
	return stripPrefix(in, prefix, c.config.tokenSeparator, c.config.keySeparator)
}

func stripPrefix(in, prefix, tokenSeparator, keySeparator string) string {
	t := in
	b := regexp.MustCompile(`[|].*`).ReplaceAll([]byte(t), []byte(""))
	return strings.Replace(string(b), fmt.Sprintf("%s%s", prefix, tokenSeparator), "", 1)
}

// FlushToFile saves contents to file provided
// in the config input into the generator
func (c *GenVars) FlushToFile() (string, error) {

	// moved up to
	joinedStr := listToString(c.outString)

	if c.config.outpath == "stdout" {
		fmt.Fprint(os.Stdout, joinedStr)
	} else {
		e := os.WriteFile(c.config.outpath, []byte(joinedStr), 0644)
		if e != nil {
			return "", e
		}
	}
	return c.config.outpath, nil
}

func listToString(strList []string) string {
	return strings.Join(strList, "\n")
}
