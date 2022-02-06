package genvars

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	klog "k8s.io/klog/v2"
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
}

type genVarsStrategy interface {
	getTokenValue(c *genVars) (s string, e error)
	setToken(s string)
}

type parsedMap map[string]string

func NewGenVars(out string, ctx context.Context) *genVars {
	defaultStrategy := NewDefatultStrategy()
	return newGenVars(defaultStrategy, out, ctx)
}

func newGenVars(e genVarsStrategy, out string, ctx context.Context) *genVars {
	return &genVars{
		implementation: e,
		ctx:            ctx,
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
	klog.Info("Strategy implementation")
	return c.implementation.getTokenValue(c)
}

func stripPrefix(in, prefix string) string {
	return strings.Replace(in, fmt.Sprintf("%s%s", prefix, TokenSeparator), "", 1)
}

func (c *genVars) Generate(tokens []string) (string, error) {

	for _, token := range tokens {
		prefix := strings.Split(token, TokenSeparator)[0]
		if found := VarPrefix[prefix]; found {
			// TODO: allow for more customization here
			rawKeyToken := strings.Split(strings.Split(token, prefix)[1], "/")
			topLevelKey := rawKeyToken[len(rawKeyToken)-1]
			rawString, err := c.implemetnationSepcificDecodedString(prefix, token)
			if err != nil {
				return "", err
			}
			trm := &parsedMap{}
			isOk := isParsed(rawString, trm)
			if isOk {
				normMap := envVarNormalize(*trm)
				c.exportVars(normMap)
			} else {
				c.exportVars(parsedMap{topLevelKey: rawString})
			}

		} else {
			fmt.Println("NotFound")
		}
	}
	if c.output == "" {
		return "", fmt.Errorf("no Tokens received that could generate an output: %v", tokens)
	}

	return c.flushToFile()
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
func isParsed(res string, trm *parsedMap) bool {
	if err := json.Unmarshal([]byte(res), &trm); err != nil {
		klog.Info("unable to parse into a k/v map returning a string instead")
		return false
	}
	return true
}

//
func envVarNormalize(pmap parsedMap) parsedMap {
	normalizedMap := parsedMap{}
	for k, v := range pmap {
		normalizedMap[normalizeKey(k)] = v
	}
	return normalizedMap
}

func (c *genVars) exportVars(exportMap parsedMap) {
	// path, err := os.Getwd()
	// if err != nil {
	// 	return "", nil
	// }
	// filePath := path + "/app.env"

	for k, v := range exportMap {
		// NOTE: \n lineending is not totaly cross platform
		c.output += fmt.Sprintf("export %s=%s\n", normalizeKey(k), v)
	}
}

func normalizeKey(k string) string {
	// TODO: include a more complete regex of vaues to replace
	// r := regexp.MustCompile(``)
	return strings.ReplaceAll(strings.ToUpper(k), "-", "")
}

func (c *genVars) flushToFile() (string, error) {

	e := os.WriteFile(c.config.Outpath, []byte(c.output), 0644)
	if e != nil {
		return "", e
	}
	return c.config.Outpath, nil
}
