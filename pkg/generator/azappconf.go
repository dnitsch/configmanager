/**
 * Azure App Config implementation
**/
package generator

import (
	"context"
	"fmt"

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
	config TokenConfigVars
}

// NewAzTableStore
func NewAzAppConf(ctx context.Context, token string, conf GenVarsConfig) (*AzAppConf, error) {

	ct := conf.ParseTokenVars(token)

	backingStore := &AzAppConf{
		ctx:    ctx,
		config: ct,
	}

	srvInit := azServiceFromToken(stripPrefix(ct.Token, AzAppConfigPrefix, conf.TokenSeparator(), conf.KeySeparator()), "https://%s.azconfig.io", 1)
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

// tokenVal in AZ table storage if an Entity contains the `value` property
// we attempt to extract it and return.
//
// From this point then normal rules of configmanager apply,
// including keySeperator and lookup.
func (imp *AzAppConf) tokenVal(v *retrieveStrategy) (string, error) {
	log.Info("Concrete implementation AzAppConf")
	log.Infof("AzAppConf Token: %s", imp.token)

	_, cancel := context.WithCancel(imp.ctx)
	defer cancel()

	return "", nil
}
