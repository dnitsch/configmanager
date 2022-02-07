package generator

import "github.com/dnitsch/genvars/pkg/log"

type DefaultStrategy struct {
}

func NewDefatultStrategy() *DefaultStrategy {
	return &DefaultStrategy{}
}

func (implmt *DefaultStrategy) setToken(token string) {
}

func (implmt *DefaultStrategy) getTokenValue(v *genVars) (string, error) {
	log.Infof("%s", "Concrete implementation Default On Startup")
	return "", nil
}
