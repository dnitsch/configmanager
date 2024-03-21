/**
 * Azure App Config implementation
**/
package generator

import (
	"context"
	"fmt"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/data/azappconfig"
	"github.com/dnitsch/configmanager/pkg/log"
)

// appConfApi
// uses this package https://pkg.go.dev/github.com/Azure/azure-sdk-for-go/sdk/data/azappconfig
type appConfApi interface {
	GetSetting(ctx context.Context, key string, options *azappconfig.GetSettingOptions) (azappconfig.GetSettingResponse, error)
}

type AzAppConf struct {
	svc    appConfApi
	ctx    context.Context
	token  string
	config *AzAppConfConfig
}

// AzAppConfConfig is the azure conf service specific config
// and it is parsed from the token metadata
type AzAppConfConfig struct {
	Label          string       `json:"label"`
	Etag           *azcore.ETag `json:"etag"`
	AcceptDateTime *time.Time   `json:"acceptedDateTime"`
}

// NewAzAppConf
func NewAzAppConf(ctx context.Context, token string, conf GenVarsConfig) (*AzAppConf, error) {
	storeConf := &AzAppConfConfig{}
	initialToken := ParseMetadata(token, storeConf)
	backingStore := &AzAppConf{
		ctx:    ctx,
		config: storeConf,
	}

	srvInit := azServiceFromToken(stripPrefix(initialToken, AzAppConfigPrefix, conf.TokenSeparator(), conf.KeySeparator()), "https://%s.azconfig.io", 1)
	backingStore.token = srvInit.token

	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	c, err := azappconfig.NewClient(srvInit.serviceUri, cred, nil)
	if err != nil {
		log.Error(err)
		return nil, fmt.Errorf("%v\n%w", err, ErrClientInitialization)
	}

	backingStore.svc = c
	return backingStore, nil

}

// setTokenVal sets the token
func (implmt *AzAppConf) setTokenVal(token string) {}

// tokenVal in AZ App Config
// label can be specified
// From this point then normal rules of configmanager apply,
// including keySeperator and lookup.
func (imp *AzAppConf) tokenVal(v *retrieveStrategy) (string, error) {
	log.Info("Concrete implementation AzAppConf")
	log.Infof("AzAppConf Token: %s", imp.token)

	ctx, cancel := context.WithCancel(imp.ctx)
	defer cancel()
	opts := &azappconfig.GetSettingOptions{}

	// assign any metadatas from the token
	if imp.config.Label != "" {
		opts.Label = &imp.config.Label
	}

	if imp.config.Etag != nil {
		opts.OnlyIfChanged = imp.config.Etag
	}

	s, err := imp.svc.GetSetting(ctx, imp.token, opts)
	if err != nil {
		log.Errorf(implementationNetworkErr, AzAppConfigPrefix, err, imp.token)
		return "", fmt.Errorf("token: %s, error: %v. %w", imp.token, err, ErrRetrieveFailed)
	}
	if s.Value != nil {
		return *s.Value, nil
	}
	log.Errorf("token: %v, %w", imp.token, ErrEmptyResponse)
	return "", nil
}
