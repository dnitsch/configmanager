package genvars

import (
	klog "k8s.io/klog/v2"
)

type DefaultStrategy struct {
}

func NewDefatultStrategy() *DefaultStrategy {
	return &DefaultStrategy{}
}

func (implmt *DefaultStrategy) setToken(token string) {
}

func (implmt *DefaultStrategy) getTokenValue(v *genVars) (string, error) {
	klog.Infof("%s", "Concrete implementation Default On Startup")
	return "", nil
}
