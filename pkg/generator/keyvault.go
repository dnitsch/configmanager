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

	// "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/keyvault/armkeyvault"
	// azkv "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/keyvault/armkeyvault"

	"github.com/dnitsch/configmanager/pkg/log"
)

type kvApi interface {
	GetSecret(ctx context.Context, name string, version string, options *azsecrets.GetSecretOptions) (azsecrets.GetSecretResponse, error)
}

type KvScrtStore struct {
	svc   kvApi
	token string
}

// azVaultHelper provides a broken up string
type azVaultHelper struct {
	vaultUri string
	token    string
}

// NewKvScrtStore returns a KvStore
// requires `AZURE_SUBSCRIPTION_ID` environment variable to be present to successfuly work
func NewKvScrtStore(ctx context.Context) (*KvScrtStore, error) {
	return &KvScrtStore{}, nil
}

// NewKvStore returns a KvStore
// requires `AZURE_SUBSCRIPTION_ID` environment variable to be present to successfuly work
func NewKvScrtStoreWithToken(ctx context.Context, token, tokenSeparator, keySeparator string) (*KvScrtStore, error) {

	//
	conf := azSplitToken(stripPrefix(token, AzKeyVaultSecretsPrefix, tokenSeparator, keySeparator))

	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	c := azsecrets.NewClient(conf.vaultUri, cred, nil)

	return &KvScrtStore{
		svc:   c,
		token: conf.token,
	}, nil
}

func (paramStr *KvScrtStore) setToken(token string) {
	paramStr.token = token
}

func (implmt *KvScrtStore) setValue(val string) {
}

func (imp *KvScrtStore) getTokenValue(v *GenVars) (string, error) {
	log.Infof("%s", "Concrete implementation AzKeyVault Secret")
	log.Infof("AzKeyVault Token: %s", imp.token)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// secretVersion as "" => latest
	s, err := imp.svc.GetSecret(ctx, imp.token, "", nil)
	if err != nil {
		log.Errorf("AzKeyVault: %s", err)
		return "", err
	}
	return *s.Value, nil
}

func azSplitToken(token string) azVaultHelper {
	// ensure preceeding slash is trimmed
	splitToken := strings.Split(strings.TrimPrefix(token, "/"), "/")
	vaultUri := fmt.Sprintf("https://%s.vault.azure.net", splitToken[0])
	return azVaultHelper{vaultUri: vaultUri, token: strings.Join(splitToken[1:], "/")}
}
