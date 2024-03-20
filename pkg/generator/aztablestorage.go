/**
 * Azure TableStore implementation
**/
package generator

import (
	"context"
	"encoding/json"
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

// NewAzTableStore
func NewAzTableStore(ctx context.Context, token string, conf GenVarsConfig) (*AzTableStore, error) {

	ct := conf.ParseTokenVars(token)

	backingStore := &AzTableStore{
		ctx:    ctx,
		config: ct,
	}

	srvInit := azServiceFromToken(stripPrefix(ct.Token, AzTableStorePrefix, conf.TokenSeparator(), conf.KeySeparator()), "https://%s.table.core.windows.net/%s", 2)
	backingStore.token = srvInit.token

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

	backingStore.svc = c
	return backingStore, nil

}

// setToken already happens in the constructor
func (implmt *AzTableStore) setTokenVal(token string) {}

// tokenVal in AZ table storage if an Entity contains the `value` property
// we attempt to extract it and return.
//
// From this point then normal rules of configmanager apply,
// including keySeperator and lookup.
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
		log.Errorf(implementationNetworkErr, AzTableStorePrefix, err, imp.token)
		return "", fmt.Errorf(implementationNetworkErr+" %w", AzTableStorePrefix, err, imp.token, ErrRetrieveFailed)
	}
	if len(s.Value) > 0 {
		// check for `value` property in entity
		checkVal := make(map[string]interface{})
		json.Unmarshal(s.Value, &checkVal)
		if checkVal["value"] != nil {
			return fmt.Sprintf("%v", checkVal["value"]), nil
		}
		return string(s.Value), nil
	}
	log.Errorf("value retrieved but empty for token: %v", imp.token)
	return "", nil
}

func azTableStoreTokenSplitter(token string) (partitionKey, rowKey string, err error) {
	splitToken := strings.Split(strings.TrimPrefix(token, "/"), "/")
	if len(splitToken) < 2 {
		return "", "", fmt.Errorf("token: %s - could not be correctly destructured to pluck the partition and row keys\n%w", token, ErrIncorrectlyStructuredToken)
	}
	partitionKey = splitToken[0]
	rowKey = splitToken[1]
	// naked return to save having to define another struct
	return
}

