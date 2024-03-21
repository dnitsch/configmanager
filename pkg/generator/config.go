package generator

import (
	"encoding/json"
	"fmt"
	"strings"
)

// TokenConfigVars
type TokenConfigVars struct {
	Token string
	// AWS IAM Role for Vault AWS IAM Auth
	Role string
	// where supported a version of the secret can be specified
	//
	// e.g. HashiVault or AWS SecretsManager or AzAppConfig Label
	//
	Version string
}

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

// ParseMetadata parses the metadata of each token into the provided pointer.
// All data inside the `[` `]` is considered metadata about the token.
//
// returns the token without the metadata `[` `]`
//
// Further processing down the line will remove other elements of the token.
func ParseMetadata[T comparable](token string, typ T) string {
	metadataStr, startIndex, found := extractMetadataStr(token)
	if !found {
		return token
	}
	if startIndex > 0 {
		token = token[0:startIndex]
	}
	// crude json like builder from key/val tags
	// since we are only ever dealing with a string input
	// extracted from the token there is little chance panic would occur here
	// WATCH THIS SPACE "¯\_(ツ)_/¯"
	metaMap := []string{}
	for _, keyVal := range strings.Split(metadataStr, ",") {
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
		return token
	}
	return token
}

const startMetaStr string = `[`
const endMetaStr string = `]`

// extractMetadataStr returns anything between the start and end
// metadata markers in the token string itself
func extractMetadataStr(str string) (metaString string, startIndex int, found bool) {

	startIndex = strings.Index(str, startMetaStr)
	if startIndex == -1 {
		return metaString, startIndex, false
	}
	newS := str[startIndex+len(startMetaStr):]
	e := strings.Index(newS, endMetaStr)
	if e == -1 {
		return metaString, -1, false
	}
	metaString = newS[:e]
	return metaString, startIndex, true
}
