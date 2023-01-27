package generator

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
