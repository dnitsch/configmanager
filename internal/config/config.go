package config

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
)

const (
	SELF_NAME = "configmanager"
)

const (
	// tokenSeparator used for identifying the end of a prefix and beginning of token
	// see notes about special consideration for AZKVSECRET tokens
	tokenSeparator = "://"
	// keySeparator used for accessing nested objects within the retrieved map
	keySeparator = "|"
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

// GenVarsConfig defines the input config object to be passed
type GenVarsConfig struct {
	outpath        string
	tokenSeparator string
	keySeparator   string
	// parseAdditionalVars func(token string) TokenConfigVars
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

// OutputPath returns the outpath set in the config
func (c *GenVarsConfig) OutputPath() string {
	return c.outpath
}

// TokenSeparator returns the tokenSeparator set in the config
func (c *GenVarsConfig) TokenSeparator() string {
	return c.tokenSeparator
}

// KeySeparator returns the keySeparator set in the config
func (c *GenVarsConfig) KeySeparator() string {
	return c.keySeparator
}

// Config returns the derefed value
func (c *GenVarsConfig) Config() GenVarsConfig {
	return *c
}

type ParsedTokenConfig struct {
	prefix                       ImplementationPrefix
	fullToken                    string
	metadataStr                  string
	tokenWithoutMetadata         string
	keysPath                     string
	keySeparator, tokenSeparator string
}

// NewParsedTokenConfig returns a pointer to a new TokenConfig struct
func NewParsedTokenConfig(token string, config GenVarsConfig) *ParsedTokenConfig {
	ptc := &ParsedTokenConfig{}
	prfx := strings.Split(token, config.TokenSeparator())[0]
	// This should already only be a list of properly supported tokens but just in case
	if found := VarPrefix[ImplementationPrefix(prfx)]; found {
		ptc.prefix = ImplementationPrefix(prfx)
		ptc.fullToken = token
		ptc.keySeparator = config.keySeparator
		ptc.tokenSeparator = config.tokenSeparator
		ptc.metadataStr, ptc.tokenWithoutMetadata = extractMetadataStr(token)
		return ptc
	}
	return nil
}

func (t *ParsedTokenConfig) ParseMetadata(typ any) error {
	// crude json like builder from key/val tags
	// since we are only ever dealing with a string input
	// extracted from the token there is little chance panic would occur here
	// WATCH THIS SPACE "¯\_(ツ)_/¯"
	metaMap := []string{}
	for _, keyVal := range strings.Split(t.metadataStr, ",") {
		mapKeyVal := strings.Split(keyVal, "=")
		if len(mapKeyVal) == 2 {
			metaMap = append(metaMap, fmt.Sprintf(`"%s":"%s"`, mapKeyVal[0], mapKeyVal[1]))
		}
	}

	// empty map will be parsed as `{}` still resulting in a valid json
	// and successful unmarshalling but default value pointer struct
	b := []byte(fmt.Sprintf(`{%s}`, strings.Join(metaMap, ",")))
	if err := json.Unmarshal(b, typ); err != nil {
		// It would very hard to test this since
		// we are forcing the key and value to be strings
		// return non-filled pointer
		return err
	}
	return nil
}

func (t *ParsedTokenConfig) StripPrefix() string {
	return stripPrefix(t.fullToken, t.prefix, t.tokenSeparator, t.keySeparator)
}

// StripMetadata returns the fullToken without the
// metadata
func (t *ParsedTokenConfig) StripMetadata() string {
	return t.tokenWithoutMetadata
}

// Full returns the full Token path.
// Including key separator and metadata values
func (t *ParsedTokenConfig) String() string {
	return t.fullToken
}

func (t *ParsedTokenConfig) Prefix() ImplementationPrefix {
	return t.prefix
}

const (
	startMetaStr string = `[`
	endMetaStr   string = `]`
)

// extractMetadataStr returns anything between the start and end
// metadata markers in the token string itself
func extractMetadataStr(token string) (metaString string, tokenWithoutMeta string) {

	startIndex := strings.Index(token, startMetaStr)
	// token has no startMetaStr
	if startIndex == -1 {
		return metaString, token
	}
	newS := token[startIndex+len(startMetaStr):]

	endIndex := strings.Index(newS, endMetaStr)
	// token has no meta end
	if endIndex == -1 {
		return metaString, token
	}
	// metastring extracted
	metaString = newS[:endIndex]

	// complete [key=value] has been found
	// Remove from the token
	metaLessToken := strings.Replace(token, startMetaStr+metaString+endMetaStr, "", -1)

	return metaString, metaLessToken
}

// stripPrefix
func stripPrefix(in string, prefix ImplementationPrefix, tokenSeparator, keySeparator string) string {
	t := in
	b := regexp.MustCompile(fmt.Sprintf(`[%s].*`, keySeparator)).ReplaceAllString(t, "")
	return strings.Replace(b, fmt.Sprintf("%s%s", prefix, tokenSeparator), "", 1)
}
