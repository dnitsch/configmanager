package generator

import (
	"context"
	"errors"
	"fmt"

	"github.com/dnitsch/configmanager/internal/config"
	"github.com/dnitsch/configmanager/internal/store"
)

var ErrTokenInvalid = errors.New("invalid token - cannot get prefix")

type RetrieveStrategy struct {
	implementation store.Strategy
	config         config.GenVarsConfig
	token          string
}

// NewRetrieveStrategy
func NewRetrieveStrategy(s store.Strategy, config config.GenVarsConfig) *RetrieveStrategy {
	return &RetrieveStrategy{implementation: s, config: config}
}

func (rs *RetrieveStrategy) setImplementation(strategy store.Strategy) {
	rs.implementation = strategy
}

func (rs *RetrieveStrategy) setTokenVal(s *config.ParsedTokenConfig) {
	rs.implementation.SetToken(s)
}

func (rs *RetrieveStrategy) getTokenValue() (string, error) {
	return rs.implementation.Token()
}

type TokenResponse struct {
	value string
	key   *config.ParsedTokenConfig
	Err   error
}

// retrieveSpecificCh wraps around the specific strategy implementation
// and publishes results to a channel
func (rs *RetrieveStrategy) RetrieveByToken(ctx context.Context, impl store.Strategy, tokenConf *config.ParsedTokenConfig) *TokenResponse {
	cr := &TokenResponse{}
	cr.Err = nil
	cr.key = tokenConf
	rs.setImplementation(impl)
	rs.setTokenVal(tokenConf)
	s, err := rs.getTokenValue()
	if err != nil {
		cr.Err = err
		return cr
	}
	cr.value = s
	return cr
}

func (rs *RetrieveStrategy) SelectImplementation(ctx context.Context, token *config.ParsedTokenConfig) (store.Strategy, error) {
	if token == nil {
		return nil, fmt.Errorf("unable to get prefix, %w", ErrTokenInvalid)
	}

	switch token.Prefix() {
	case config.AzTableStorePrefix:
		return store.NewAzTableStore(ctx, token)
	// case SecretMgrPrefix:
	// 	return NewSecretsMgr(ctx)
	// case ParamStorePrefix:
	// 	return NewParamStore(ctx)
	// case AzKeyVaultSecretsPrefix:
	// 	return NewKvScrtStore(ctx, in, config)
	// case GcpSecretsPrefix:
	// 	return NewGcpSecrets(ctx)
	// case HashicorpVaultPrefix:
	// 	return NewVaultStore(ctx, in, config)
	case config.AzAppConfigPrefix:
		return store.NewAzAppConf(ctx, token)
	default:
		return nil, fmt.Errorf("implementation not found for input string: %s", token)
	}
}
