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
	TokenSeparator   = "#"
	SecretMgrPrefix  = "AWSSECRETS"
	ParamStorePrefix = "AWSPARAMSTR"
)

var (
	VarPrefix = map[string]bool{SecretMgrPrefix: true, ParamStorePrefix: true}
)

// GenVarsConfig defines the input config object to be passed
type GenVarsConfig struct {
	Outpath        string
	TokenSeparator string
}

type genVars struct {
	implementation genVarsStrategy
	token          string
	ctx            context.Context
	config         GenVarsConfig
	outString      string
	// rawMap is the internal object that holds the values of original token => retrieved value
	rawMap ParsedMap
}

type genVarsStrategy interface {
	getTokenValue(c *genVars) (s string, e error)
	setToken(s string)
	setValue(s string)
}

// ParsedMap is the internal working object definition and
// the return type if results are not flushed to file
type ParsedMap map[string]interface{}

func New() *genVars {
	defaultStrategy := NewDefatultStrategy()
	return newGenVars(defaultStrategy)
}

func newGenVars(e genVarsStrategy) *genVars {
	m := ParsedMap{}
	return &genVars{
		implementation: e,
		rawMap:         m,
	}
}

func (c *genVars) WithConfig(cfg *GenVarsConfig) *genVars {
	if cfg.TokenSeparator == "" {
		cfg.TokenSeparator = TokenSeparator
	}
	c.config = *cfg
	return c
}

func (c *genVars) setImplementation(strategy genVarsStrategy) {
	c.implementation = strategy
}

func (c *genVars) setToken(s string) {
	c.implementation.setToken(s)
}

func (c *genVars) setVaule(s string) {
	c.implementation.setValue(s)
}

func (c *genVars) getTokenValue() (string, error) {
	log.Info("Strategy implementation")
	return c.implementation.getTokenValue(c)
}

func (c *genVars) stripPrefix(in, prefix string) string {
	return strings.Replace(in, fmt.Sprintf("%s%s", prefix, c.config.TokenSeparator), "", 1)
}

// Generate generates a k/v map of the tokens with their corresponding secret/paramstore values
// the standard pattern of a token should follow a path like
func (c *genVars) Generate(tokens []string) (ParsedMap, error) {

	for _, token := range tokens {
		prefix := strings.Split(token, TokenSeparator)[0]
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

func (c *genVars) retrieveSpecific(prefix, in string) (string, error) {
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
		return "", fmt.Errorf("ImplementationNotFound for input string: %s", in)
	}
}

// isParsed will try to parse the return found string into
// map[string]string
// If found it will convert that to a map with all keys uppercased
// and any characters
func isParsed(res interface{}, trm *ParsedMap) bool {
	str := fmt.Sprint(res)
	if err := json.Unmarshal([]byte(str), &trm); err != nil {
	    log.Infof("Err = %v", err)
		log.Info("unable to parse into a k/v map returning a string instead")
		return false
	}
	return true
}

// ConvertToExportVar
func (c *genVars) ConvertToExportVar() {
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

func (c *genVars) exportVars(exportMap ParsedMap) {

	for k, v := range exportMap {
		// NOTE: \n line ending is not totaly cross platform
		_type := fmt.Sprintf("%T", v)
		switch _type {
		case "string":
			c.outString += fmt.Sprintf("export %s='%s'\n", normalizeKey(k), v)
		default:
			c.outString += fmt.Sprintf("export %s=%v\n", normalizeKey(k), v)
		}

	}
	// c.mapOut[normalizeKey(k)] = v
}

func normalizeKey(k string) string {
	// TODO: include a more complete regex of vaues to replace
	r := regexp.MustCompile(`[\s\@\!]`).ReplaceAll([]byte(k), []byte(""))
	r = regexp.MustCompile(`[\-]`).ReplaceAll(r, []byte("_"))
	return strings.ToUpper(string(r))
}

func (c *genVars) FlushToFile() (string, error) {

	e := os.WriteFile(c.config.Outpath, []byte(c.outString), 0644)
	if e != nil {
		return "", e
	}
	return c.config.Outpath, nil
}
