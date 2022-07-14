package generator

type genVarsStrategy interface {
	getTokenValue(c *GenVars) (s string, e error)
	setToken(s string)
	setValue(s string)
}

func (c *GenVars) setImplementation(strategy genVarsStrategy) {
	c.implementation = strategy
}

func (c *GenVars) setToken(s string) {
	c.implementation.setToken(s)
}

func (c *GenVars) setVaule(s string) {
	c.implementation.setValue(s)
}

func (c *GenVars) getTokenValue() (string, error) {
	return c.implementation.getTokenValue(c)
}
