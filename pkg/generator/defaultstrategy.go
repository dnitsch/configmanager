package generator

import "fmt"

type DefaultStrategy struct {
}

const implementationNetworkErr string = "implementation %s error: %v for token: %s"

func NewDefatultStrategy() *DefaultStrategy {
	return &DefaultStrategy{}
}

// setToken on default strategy
func (implmt *DefaultStrategy) setToken(token string) {}

// setValue on default strategy
func (implmt *DefaultStrategy) setValue(val string) {}

func (implmt *DefaultStrategy) tokenVal(v *retrieveStrategy) (string, error) {
	return "", fmt.Errorf("default strategy does not implement token retrieval")
}
