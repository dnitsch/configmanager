/**
 * Azure KeyVault implementation
**/
package store

import (
	"context"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/security/keyvault/azsecrets"
	"github.com/dnitsch/configmanager/internal/config"
	"github.com/dnitsch/configmanager/pkg/log"
)

type kvApi interface {
	GetSecret(ctx context.Context, name string, version string, options *azsecrets.GetSecretOptions) (azsecrets.GetSecretResponse, error)
}

type KvScrtStore struct {
	svc           kvApi
	ctx           context.Context
	token         *config.ParsedTokenConfig
	config        *AzKvConfig
	strippedToken string
}

// AzKvConfig takes any metadata from the token
// Version is the only
type AzKvConfig struct {
	Version string `json:"version"`
}

// NewKvScrtStore returns a KvScrtStore
// requires `AZURE_SUBSCRIPTION_ID` environment variable to be present to successfully work
func NewKvScrtStore(ctx context.Context, token *config.ParsedTokenConfig) (*KvScrtStore, error) {

	storeConf := &AzKvConfig{}
	token.ParseMetadata(storeConf)

	backingStore := &KvScrtStore{
		ctx:    ctx,
		config: storeConf,
		token:  token,
	}

	srvInit := azServiceFromToken(token.StoreToken(), "https://%s.vault.azure.net", 1)
	backingStore.strippedToken = srvInit.token

	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	c, err := azsecrets.NewClient(srvInit.serviceUri, cred, nil)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	backingStore.svc = c
	return backingStore, nil

}

// setToken already happens in AzureKVClient in the constructor
func (implmt *KvScrtStore) SetToken(token *config.ParsedTokenConfig) {}

func (imp *KvScrtStore) Token() (string, error) {
	log.Info("Concrete implementation AzKeyVault Secret")
	log.Infof("AzKeyVault Token: %s", imp.token.String())

	ctx, cancel := context.WithCancel(imp.ctx)
	defer cancel()

	// secretVersion as "" => latest
	// imp.config.Version will default `""` if not specified
	s, err := imp.svc.GetSecret(ctx, imp.strippedToken, imp.config.Version, nil)
	if err != nil {
		log.Errorf(implementationNetworkErr, imp.token.Prefix(), err, imp.token.String())
		return "", err
	}
	if s.Value != nil {
		return *s.Value, nil
	}
	log.Errorf("value retrieved but empty for token: %v", imp.token)
	return "", nil
}
