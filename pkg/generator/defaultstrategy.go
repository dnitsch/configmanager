package generator

import (
	"errors"
	"fmt"
)

type DefaultStrategy struct {
}

const implementationNetworkErr string = "implementation %s error: %v for token: %s"

var (
	ErrEmptyResponse     = errors.New("value retrieved but empty for token")
	ErrServiceCallFailed = errors.New("failed to complete the service call")
)

func NewDefatultStrategy() *DefaultStrategy {
	return &DefaultStrategy{}
}

// setToken on default strategy
func (implmt *DefaultStrategy) setTokenVal(token string) {}

// setValue on default strategy
func (implmt *DefaultStrategy) setValue(val string) {}

func (implmt *DefaultStrategy) tokenVal(v *retrieveStrategy) (string, error) {
	return "", fmt.Errorf("default strategy does not implement token retrieval")
}
