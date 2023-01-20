package generator

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/dnitsch/configmanager/pkg/log"

	vault "github.com/hashicorp/vault/api"
)

// vaultHelper provides a broken up string
type vaultHelper struct {
	path  string
	token string
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
	secToken, found := os.LookupEnv("VAULT_TOKEN")
	if !found {
		return nil, fmt.Errorf("VAULT_TOKEN not specified, cannot initialize vault client")
	}
	// Authenticate
	client.SetToken(secToken)

	return &VaultStore{
		svc:   client.KVv2(vt.path),
		ctx:   ctx,
		token: vt.token,
	}, nil
}

// setToken already happens in Vault constructor
// no need to re-set it here
func (implmt *VaultStore) setToken(token string) {
}

func (implmt *VaultStore) setValue(val string) {
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
		log.Errorf("unable to read secret: %v", err)
		return "", err
	}

	if secret.Data != nil {
		return marshall(secret.Data)
	}

	log.Errorf("value retrieved but empty for token: %v", imp.token)
	return "", nil
}

// // verifyVaultToken ensures the token includes
// // a key separator to perform a lookup on the
// // secret KV
// func verifyVaultToken(token, keySeparator string) (vaultHelper, error) {
// 	vh := vaultHelper{}
// 	s := strings.Split(token, keySeparator)
// 	if len(s[1]) < 1 {
// 		return vh, fmt.Errorf("vault needs a key specified on each token in order to be able to retrieve it")
// 	}
// 	vh.key = s[1]
// 	return splitToken(token, vh), nil
// 	// s := strings.Split(strings.TrimPrefix(token, "/"), "/")
// 	// return vaultHelper{token: strings.Join(s[1:], "/"), path: s[0]}
// }

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
