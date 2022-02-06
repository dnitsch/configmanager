package genvars

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/dnitsch/genvars/pkg/log"
)

const (
	TokenSeparator   = "#"
	SecretMgrPrefix  = "AWSSECRETS"
	ParamStorePrefix = "AWSPARAMSTR"
)

var (
	VarPrefix = map[string]bool{SecretMgrPrefix: true, ParamStorePrefix: true}
)

type GenVarsConfig struct {
	Outpath        string
	TokenSeparator string
}

type genVars struct {
	implementation genVarsStrategy
	token          string
	ctx            context.Context
	config         GenVarsConfig
	output         string
	mapOut         ParsedMap
}

type genVarsStrategy interface {
	getTokenValue(c *genVars) (s string, e error)
	setToken(s string)
}

type ParsedMap map[string]string

func NewGenVars(out string, ctx context.Context) *genVars {
	defaultStrategy := NewDefatultStrategy()
	return newGenVars(defaultStrategy, out, ctx)
}

func newGenVars(e genVarsStrategy, out string, ctx context.Context) *genVars {
	mo := ParsedMap{}
	return &genVars{
		implementation: e,
		ctx:            ctx,
		mapOut:         mo,
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

func (c *genVars) getTokenValue() (string, error) {
	log.Info("Strategy implementation")
	return c.implementation.getTokenValue(c)
}

func (c *genVars) stripPrefix(in, prefix string) string {
	return strings.Replace(in, fmt.Sprintf("%s%s", prefix, c.config.TokenSeparator), "", 1)
}

func (c *genVars) Generate(tokens []string) (ParsedMap, error) {

	for _, token := range tokens {
		prefix := strings.Split(token, TokenSeparator)[0]
		if found := VarPrefix[prefix]; found {
			// TODO: allow for more customization here
			rawKeyToken := strings.Split(strings.Split(token, prefix)[1], "/")
			topLevelKey := rawKeyToken[len(rawKeyToken)-1]
			rawString, err := c.implemetnationSepcificDecodedString(prefix, token)
			if err != nil {
				return nil, err
			}
			trm := &ParsedMap{}
			isOk := isParsed(rawString, trm)
			if isOk {
				normMap := envVarNormalize(*trm)
				c.exportVars(normMap)
			} else {
				c.exportVars(ParsedMap{topLevelKey: rawString})
			}

		} else {
			log.Info("NotFound")
		}
	}
	if c.output == "" {
		return nil, fmt.Errorf("no Tokens received that could generate an output: %v", tokens)
	}

	return c.mapOut, nil
	// return c.flushToFile()
}

func (c *genVars) implemetnationSepcificDecodedString(prefix, in string) (string, error) {
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
func isParsed(res string, trm *ParsedMap) bool {
	if err := json.Unmarshal([]byte(res), &trm); err != nil {
		log.Info("unable to parse into a k/v map returning a string instead")
		return false
	}
	return true
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
		// NOTE: \n lineending is not totaly cross platform
		c.output += fmt.Sprintf("export %s='%s'\n", normalizeKey(k), v)
		c.mapOut[normalizeKey(k)] = v
	}
}

func normalizeKey(k string) string {
	// TODO: include a more complete regex of vaues to replace
	// r := regexp.MustCompile(``)
	return strings.ReplaceAll(strings.ToUpper(k), "-", "")
}

func (c *genVars) FlushToFile() (string, error) {

	e := os.WriteFile(c.config.Outpath, []byte(c.output), 0644)
	if e != nil {
		return "", e
	}
	return c.config.Outpath, nil
}
