/**
 * Azure KeyVault implementation
**/
package generator

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/data/aztables"
	"github.com/dnitsch/configmanager/pkg/log"
)

var ErrIncorrectlyStructuredToken = errors.New("incorrectly structured token")

// tableStoreApi
// uses this package https://pkg.go.dev/github.com/Azure/azure-sdk-for-go/sdk/data/aztables
type tableStoreApi interface {
	GetEntity(ctx context.Context, partitionKey string, rowKey string, options *aztables.GetEntityOptions) (aztables.GetEntityResponse, error)
}

type AzTableStore struct {
	svc    tableStoreApi
	ctx    context.Context
	token  string
	config TokenConfigVars
}

// NewAzTableStore returns a KvScrtStore
// requires `AZURE_SUBSCRIPTION_ID` environment variable to be present to successfully work
func NewAzTableStore(ctx context.Context, token string, conf GenVarsConfig) (*AzTableStore, error) {

	ct := conf.ParseTokenVars(token)

	tstore := &AzTableStore{
		ctx:    ctx,
		config: ct,
	}

	srvInit := azServiceFromToken(stripPrefix(ct.Token, AzTableStorePrefix, conf.TokenSeparator(), conf.KeySeparator()), "https://%s.table.core.windows.net/%s", 2)
	tstore.token = srvInit.token

	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	c, err := aztables.NewClient(srvInit.serviceUri, cred, nil)
	if err != nil {
		log.Error(err)
		return nil, fmt.Errorf("%v\n%w", err, ErrClientInitialization)
	}

	tstore.svc = c
	return tstore, nil

}

// setToken already happens in the constructor
func (implmt *AzTableStore) setToken(token string) {}

func (imp *AzTableStore) tokenVal(v *retrieveStrategy) (string, error) {
	log.Info("Concrete implementation AzTableSTore")
	log.Infof("AzTableSTore Token: %s", imp.token)

	ctx, cancel := context.WithCancel(imp.ctx)
	defer cancel()

	// split the token for partition and rowKey
	pKey, rKey, err := azTableStoreTokenSplitter(imp.token)
	if err != nil {
		return "", err
	}

	s, err := imp.svc.GetEntity(ctx, pKey, rKey, &aztables.GetEntityOptions{})
	if err != nil {
		log.Errorf(implementationNetworkErr, AzKeyVaultSecretsPrefix, err, imp.token)
		return "", fmt.Errorf(implementationNetworkErr+" %w", AzKeyVaultSecretsPrefix, err, imp.token, ErrRetrieveFailed)
	}
	if s.Value != nil {
		return string(s.Value), nil
	}
	log.Errorf("value retrieved but empty for token: %v", imp.token)
	return "", nil
}

func azTableStoreTokenSplitter(token string) (
	partitionKey, rowKey string, err error) {
	splitToken := strings.Split(strings.TrimPrefix(token, "/"), "/")
	if len(splitToken) < 2 {
		return "", "", fmt.Errorf("token: %s - could not be correctly destructured to pluck the partition and row keys\n%w", token, ErrIncorrectlyStructuredToken)
	}
	partitionKey = splitToken[0]
	rowKey = splitToken[1]
	// naked return to save having to define another struct
	return
}

// Generic Azure Service Init Helpers
//
// azTableStoreHelper returns a service URI and the stripped token
type azServiceHelper struct {
	serviceUri string
	token      string
}

// azServiceFromToken for azure the first part of the token __must__ always be the
// identifier of the service e.g. the account name for tableStore or the KV name for KVSecret
// take parameter specifies the number of elements to take from the start only
//
// e.g. a value of 2 for take  will take first 2 elements from the slices
func azServiceFromToken(token string, formatUri string, take int) azServiceHelper {
	// ensure preceding slash is trimmed
	stringToken := strings.Split(strings.TrimPrefix(token, "/"), "/")
	splitToken := []any{}
	// recast []string slice to an []any
	for _, st := range stringToken {
		splitToken = append(splitToken, st)
	}

	uri := fmt.Sprintf(formatUri, splitToken[0:take]...)
	return azServiceHelper{serviceUri: uri, token: strings.Join(stringToken[take:], "/")}
}
