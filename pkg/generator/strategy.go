package generator

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
)

var (
	ErrRetrieveFailed       = errors.New("failed to retrieve config item")
	ErrClientInitialization = errors.New("failed to initialize the client")
)

type retrieveStrategy struct {
	implementation genVarsStrategy
	config         GenVarsConfig
	token          string
}

func newRetrieveStrategy(s genVarsStrategy, config GenVarsConfig) *retrieveStrategy {
	return &retrieveStrategy{implementation: s, config: config}
}

type genVarsStrategy interface {
	// getTokenConfig() AdditionalVars
	// setTokenConfig(AdditionalVars)
	tokenVal(rs *retrieveStrategy) (s string, e error)
	setToken(s string)
}

func (rs *retrieveStrategy) setImplementation(strategy genVarsStrategy) {
	rs.implementation = strategy
}

func (rs *retrieveStrategy) setToken(s string) {
	rs.implementation.setToken(s)
}

func (rs *retrieveStrategy) getTokenValue() (string, error) {
	return rs.implementation.tokenVal(rs)
}

// retrieveSpecificCh wraps around the specific strategy implementation
// and publishes results to a channel
func (rs *retrieveStrategy) RetrieveByToken(ctx context.Context, impl genVarsStrategy, prefix ImplementationPrefix, in string) chanResp {
	cr := chanResp{}
	cr.err = nil
	cr.key = in
	rs.setImplementation(impl)
	rs.setToken(in)
	s, err := rs.getTokenValue()
	if err != nil {
		cr.err = err
		return cr
	}
	cr.value = s
	return cr
}

func (rs *retrieveStrategy) SelectImplementation(ctx context.Context, prefix ImplementationPrefix, in string, config GenVarsConfig) (genVarsStrategy, error) {
	switch prefix {
	case SecretMgrPrefix:
		return NewSecretsMgr(ctx)
	case ParamStorePrefix:
		return NewParamStore(ctx)
	case AzKeyVaultSecretsPrefix:
		return NewKvScrtStore(ctx, in, config)
	case GcpSecretsPrefix:
		return NewGcpSecrets(ctx)
	case HashicorpVaultPrefix:
		return NewVaultStore(ctx, in, config)
	case AzTableStorePrefix:
		return NewAzTableStore(ctx, in, config)
	default:
		return nil, fmt.Errorf("implementation not found for input string: %s", in)
	}
}

// stripPrefix returns the token which the config/secret store
// expects to find in a provided vault/paramstore
func (rs *retrieveStrategy) stripPrefix(in string, prefix ImplementationPrefix) string {
	return stripPrefix(in, prefix, rs.config.tokenSeparator, rs.config.keySeparator)
}

// stripPrefix
func stripPrefix(in string, prefix ImplementationPrefix, tokenSeparator, keySeparator string) string {
	t := in
	b := regexp.MustCompile(fmt.Sprintf(`[%s].*`, keySeparator)).ReplaceAllString(t, "")
	return strings.Replace(b, fmt.Sprintf("%s%s", prefix, tokenSeparator), "", 1)
}
