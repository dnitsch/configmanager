/**
 * Azure KeyVault implementation
**/
package generator

import (
	"context"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/security/keyvault/azsecrets"
	"github.com/dnitsch/configmanager/pkg/log"
)

type kvApi interface {
	GetSecret(ctx context.Context, name string, version string, options *azsecrets.GetSecretOptions) (azsecrets.GetSecretResponse, error)
}

type KvScrtStore struct {
	svc    kvApi
	ctx    context.Context
	token  string
	config TokenConfigVars
}

// NewKvScrtStore returns a KvScrtStore
// requires `AZURE_SUBSCRIPTION_ID` environment variable to be present to successfully work
func NewKvScrtStore(ctx context.Context, token string, conf GenVarsConfig) (*KvScrtStore, error) {

	ct := conf.ParseTokenVars(token)

	kv := &KvScrtStore{
		ctx:    ctx,
		config: ct,
	}

	vc := azServiceFromToken(stripPrefix(ct.Token, AzKeyVaultSecretsPrefix, conf.TokenSeparator(), conf.KeySeparator()), "https://%s.vault.azure.net", 1)
	kv.token = vc.token

	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	c, err := azsecrets.NewClient(vc.serviceUri, cred, nil)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	kv.svc = c
	return kv, nil

}

// setToken already happens in AzureKVClient in the constructor
func (implmt *KvScrtStore) setTokenVal(token string) {}

func (imp *KvScrtStore) tokenVal(v *retrieveStrategy) (string, error) {
	log.Infof("%s", "Concrete implementation AzKeyVault Secret")
	log.Infof("AzKeyVault Token: %s", imp.token)

	ctx, cancel := context.WithCancel(imp.ctx)
	defer cancel()

	// secretVersion as "" => latest
	// imp.config.Version will default `""` if not specified
	s, err := imp.svc.GetSecret(ctx, imp.token, imp.config.Version, nil)
	if err != nil {
		log.Errorf(implementationNetworkErr, AzKeyVaultSecretsPrefix, err, imp.token)
		return "", err
	}
	if s.Value != nil {
		return *s.Value, nil
	}
	log.Errorf("value retrieved but empty for token: %v", imp.token)
	return "", nil
}
