package generator

import (
	"context"
	"fmt"
	"regexp"
	"strings"
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
	getTokenValue(rs *retrieveStrategy) (s string, e error)
	setToken(s string)
	setValue(s string)
}

func (rs *retrieveStrategy) setImplementation(strategy genVarsStrategy) {
	rs.implementation = strategy
}

func (rs *retrieveStrategy) setToken(s string) {
	rs.implementation.setToken(s)
}

func (rs *retrieveStrategy) setVaule(s string) {
	rs.implementation.setValue(s)
}

func (rs *retrieveStrategy) getTokenValue() (string, error) {
	return rs.implementation.getTokenValue(rs)
}

// retrieveSpecificCh wraps around the specific strategy implementation
// and publishes results to a provided channel
func (rs *retrieveStrategy) retrieveSpecificCh(ctx context.Context, prefix ImplementationPrefix, in string) chanResp {
	cr := chanResp{}
	cr.err = nil
	cr.key = in
	s, err := rs.retrieveSpecific(ctx, prefix, in)
	if err != nil {
		cr.err = err
		return cr
	}
	cr.value = s
	return cr
}

// retrieveSpecific executes a specif retrieval strategy
// based on the found token prefix
func (rs *retrieveStrategy) retrieveSpecific(ctx context.Context, prefix ImplementationPrefix, in string) (string, error) {
	switch prefix {
	case SecretMgrPrefix:
		// default strategy paramstore
		scrtMgr, err := NewSecretsMgr(ctx)
		if err != nil {
			return "", err
		}
		rs.setImplementation(scrtMgr)
		rs.setToken(in)
		return rs.getTokenValue()
	case ParamStorePrefix:
		paramStr, err := NewParamStore(ctx)
		if err != nil {
			return "", err
		}
		rs.setImplementation(paramStr)
		rs.setToken(in)
		return rs.getTokenValue()
	case AzKeyVaultSecretsPrefix:
		azKv, err := NewKvScrtStore(ctx, in, rs.config.tokenSeparator, rs.config.keySeparator)
		if err != nil {
			return "", err
		}
		// Need to swap this around for AzKV as the
		// client is initialised via vaultURI
		// and sets the token on the implementation init via NewSrv
		rs.setImplementation(azKv)
		return rs.getTokenValue()
	case GcpSecretsPrefix:
		gcpSecret, err := NewGcpSecrets(ctx)
		if err != nil {
			return "", err
		}
		rs.setImplementation(gcpSecret)
		rs.setToken(in)
		return rs.getTokenValue()
	case HashicorpVaultPrefix:
		vault, err := NewVaultStore(ctx, in, rs.config.tokenSeparator, rs.config.keySeparator)
		if err != nil {
			return "", err
		}
		rs.setImplementation(vault)
		return rs.getTokenValue()
	default:
		return "", fmt.Errorf("implementation not found for input string: %s", in)
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
	b := regexp.MustCompile(fmt.Sprintf(`[%s].*`, keySeparator)).ReplaceAll([]byte(t), []byte(""))
	return strings.Replace(string(b), fmt.Sprintf("%s%s", prefix, tokenSeparator), "", 1)
}
