package generator

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/dnitsch/configmanager/pkg/log"

	vault "github.com/hashicorp/vault/api"
	auth "github.com/hashicorp/vault/api/auth/aws"
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
	var client *vault.Client
	config := vault.DefaultConfig()
	vt := splitToken(stripPrefix(token, HashicorpVaultPrefix, tokenSeparator, keySeparator))
	client, err := vault.NewClient(config)
	if err != nil {
		log.Errorf("unable to initialize Vault client: %v", err)
	}

	if strings.HasPrefix(os.Getenv("VAULT_TOKEN"), "aws_iam") {
		awsclient, err := newVaultStoreWithAWSAuthIAM(client, "todo_get_from_token_or_other")
		if err != nil {
			return nil, err
		}
		client = awsclient
	}

	return &VaultStore{
		svc:   client.KVv2(vt.path),
		ctx:   ctx,
		token: vt.token,
	}, nil
}

// newVaultStoreWithAWSAuthIAM returns an initialised client with AWSIAMAuth
func newVaultStoreWithAWSAuthIAM(client *vault.Client, role string) (*vault.Client, error) {
	if len(role) < 1 {
		return nil, fmt.Errorf("role provided is empty, EC2 auth not supported")
	}
	awsAuth, err := auth.NewAWSAuth(
		auth.WithRole(role), // if not provided, Vault will fall back on looking for a role with the IAM role name if you're using the iam auth type, or the EC2 instance's AMI id if using the ec2 auth type
	)
	if err != nil {
		return nil, fmt.Errorf("unable to initialize AWS auth method: %w", err)
	}

	authInfo, err := client.Auth().Login(context.Background(), awsAuth)

	if err != nil {
		return nil, fmt.Errorf("unable to login to AWS auth method: %w", err)
	}
	if authInfo == nil {
		return nil, fmt.Errorf("no auth info was returned after login")
	}

	return client, nil
}

// setToken already happens in Vault constructor
// no need to re-set it here
func (imp *VaultStore) setToken(token string) {
	// this happens inside the New func call
	// due to the way the client needs to be
	// initialised with a mountpath
	// and mountpath is part of the token so it is set then
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
