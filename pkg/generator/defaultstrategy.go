package generator

import "fmt"

type DefaultStrategy struct {
}

const implementationNetworkErr string = "implementation %s error: %v for token: %s"

func NewDefatultStrategy() *DefaultStrategy {
	return &DefaultStrategy{}
}

func (implmt *DefaultStrategy) setToken(token string) {
	// this should never be used
}

func (implmt *DefaultStrategy) setValue(val string) {
	// this should never be used
}

func (implmt *DefaultStrategy) getTokenValue(v *retrieveStrategy) (string, error) {
	return "", fmt.Errorf("default strategy does not implement token retrieval")
}
