package generator

import "fmt"

type DefaultStrategy struct {
}

const implementationNetworkErr string = "implementation %s error: %v for token: %s"

func NewDefatultStrategy() *DefaultStrategy {
	return &DefaultStrategy{}
}

// setToken on default strategy
func (implmt *DefaultStrategy) setToken(token string) {
	// default strategy impl of setToken
}

// setValue on default strategy
func (implmt *DefaultStrategy) setValue(val string) {
	// default strategy impl of setValue
}

func (implmt *DefaultStrategy) tokenVal(v *retrieveStrategy) (string, error) {
	return "", fmt.Errorf("default strategy does not implement token retrieval")
}
