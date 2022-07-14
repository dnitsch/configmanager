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

func (c *GenVars) stripPrefix(in, prefix string) string {
	return strings.Replace(in, fmt.Sprintf("%s%s", prefix, c.config.tokenSeparator), "", 1)
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
			c.rawMap[token] = rawString
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
	default:
		return "", fmt.Errorf("implementationNotFound for input string: %s", in)
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

// ConvertToExportVar
func (c *GenVars) ConvertToExportVar() {
	for k, v := range c.rawMap {
		rawKeyToken := strings.Split(k, "/") // assumes a path like token was used
		topLevelKey := rawKeyToken[len(rawKeyToken)-1]
		trm := &ParsedMap{}
		isOk := isParsed(v, trm)
		if isOk {
			normMap := envVarNormalize(*trm)
			c.exportVars(normMap)
		} else {
			c.exportVars(ParsedMap{topLevelKey: v})
		}
	}
}

//
func envVarNormalize(pmap ParsedMap) ParsedMap {
	normalizedMap := ParsedMap{}
	for k, v := range pmap {
		normalizedMap[normalizeKey(k)] = v
	}
	return normalizedMap
}

func (c *GenVars) exportVars(exportMap ParsedMap) {

	for k, v := range exportMap {
		// NOTE: \n line ending is not totally cross platform
		_type := fmt.Sprintf("%T", v)
		switch _type {
		case "string":
			c.outString = append(c.outString, fmt.Sprintf("export %s='%s'", normalizeKey(k), v))
		default:
			c.outString = append(c.outString, fmt.Sprintf("export %s=%v", normalizeKey(k), v))
		}
	}
}

func normalizeKey(k string) string {
	// TODO: include a more complete regex of vaues to replace
	r := regexp.MustCompile(`[\s\@\!]`).ReplaceAll([]byte(k), []byte(""))
	r = regexp.MustCompile(`[\-]`).ReplaceAll(r, []byte("_"))
	return strings.ToUpper(string(r))
}

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
