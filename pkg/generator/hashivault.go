package generator

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/dnitsch/configmanager/pkg/log"

	vault "github.com/hashicorp/vault/api"
)

// vaultHelper provides a broken up string
type vaultHelper struct {
	path  string
	token string
}

type hashiVaultClient interface {
	KVv2(mountPath string) *vault.KVv2
}

type hashiVaultApi interface {
	Get(ctx context.Context, secretPath string) (*vault.KVSecret, error)
}

type VaultStore struct {
	svc   hashiVaultApi
	ctx   context.Context
	token string
}

func NewVaultStore(ctx context.Context, token, tokenSeparator, keySeparator string) (*VaultStore, error) {
	config := vault.DefaultConfig()
	vt := splitToken(stripPrefix(token, HashicorpVaultPrefix, tokenSeparator, keySeparator))
	client, err := vault.NewClient(config)
	if err != nil {
		log.Errorf("unable to initialize Vault client: %v", err)
	}

	return &VaultStore{
		svc:   client.KVv2(vt.path),
		ctx:   ctx,
		token: vt.token,
	}, nil
}

// func newVaultStore(ctx context.Context) (*VaultStore, error) {

// }

// setToken already happens in Vault constructor
// no need to re-set it here
func (implmt *VaultStore) setToken(token string) {
}

// getTokenValue implements the underlying techonology
// token retrieval and returns a stringified version
// of the secret
func (imp *VaultStore) getTokenValue(v *retrieveStrategy) (string, error) {
	log.Infof("%s", "Concrete implementation HashiVault")
	log.Infof("Getting Secret: %s", imp.token)

	ctx, cancel := context.WithCancel(imp.ctx)
	defer cancel()

	secret, err := imp.svc.Get(ctx, v.stripPrefix(imp.token, HashicorpVaultPrefix))
	if err != nil {
		log.Errorf(implementationNetworkErr, HashicorpVaultPrefix, err, imp.token)
		return "", err
	}

	if secret.Data != nil {
		resp, err := marshall(secret.Data)
		if err != nil {
			log.Errorf("marshalling error: %s", err.Error())
			return "", err
		}
		log.Debugf("marhalled kvv2: %s", resp)
		return resp, nil
	}

	log.Errorf("value retrieved but empty for token: %v", imp.token)
	return "", nil
}

func splitToken(token string) vaultHelper {
	vh := vaultHelper{}
	s := strings.Split(strings.TrimPrefix(token, "/"), "/")
	vh.token = strings.Join(s[1:], "/")
	vh.path = s[0]
	return vh
}

// marshall converts map[string]any into a JSON
// object. Secrets should only be a single level
// deep.
func marshall(secret map[string]any) (string, error) {
	b, err := json.Marshal(secret)
	if err != nil {
		return "", err
	}
	return string(b), nil
}
