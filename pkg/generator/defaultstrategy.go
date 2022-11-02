package generator

import "fmt"

type DefaultStrategy struct {
}

func NewDefatultStrategy() *DefaultStrategy {
	return &DefaultStrategy{}
}

func (implmt *DefaultStrategy) setToken(token string) {
}

func (implmt *DefaultStrategy) setValue(val string) {
}

func (implmt *DefaultStrategy) getTokenValue(v *retrieveStrategy) (string, error) {
	return "", fmt.Errorf("default strategy does not implement token retrieval")
}
