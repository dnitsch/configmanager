/**
 * Azure KeyVault implementation
**/
package generator

import (
	"context"
	"fmt"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/keyvault/azsecrets"

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

// azVaultHelper provides a broken up string
type azVaultHelper struct {
	vaultUri string
	token    string
}

// NewKvScrtStore returns a KvScrtStore
// requires `AZURE_SUBSCRIPTION_ID` environment variable to be present to successfully work
func NewKvScrtStore(ctx context.Context, token string, conf GenVarsConfig) (*KvScrtStore, error) {

	ct := conf.ParseTokenVars(token)

	kv := &KvScrtStore{
		ctx:    ctx,
		config: ct,
	}

	vc := azSplitToken(stripPrefix(ct.Token, AzKeyVaultSecretsPrefix, conf.TokenSeparator(), conf.KeySeparator()))
	kv.token = vc.token

	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	c, err := azsecrets.NewClient(vc.vaultUri, cred, nil)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	kv.svc = c
	return kv, nil

}

func (implmt *KvScrtStore) setToken(token string) {
	// setToken already happens in AzureKVClient in the constructor
	// no need to re-set it here
}

func (imp *KvScrtStore) getTokenValue(v *retrieveStrategy) (string, error) {
	log.Infof("%s", "Concrete implementation AzKeyVault Secret")
	log.Infof("AzKeyVault Token: %s", imp.token)

	ctx, cancel := context.WithCancel(imp.ctx)
	defer cancel()

	// secretVersion as "" => latest
	s, err := imp.svc.GetSecret(ctx, imp.token, "", nil)
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

func azSplitToken(token string) azVaultHelper {
	// ensure preceding slash is trimmed
	splitToken := strings.Split(strings.TrimPrefix(token, "/"), "/")
	vaultUri := fmt.Sprintf("https://%s.vault.azure.net", splitToken[0])
	return azVaultHelper{vaultUri: vaultUri, token: strings.Join(splitToken[1:], "/")}
}
