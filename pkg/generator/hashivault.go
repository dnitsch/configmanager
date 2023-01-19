package generator

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/dnitsch/configmanager/pkg/log"

	vault "github.com/hashicorp/vault/api"
)

type hashiVaultApi interface {
	GetParameter(ctx context.Context, params *ssm.GetParameterInput, optFns ...func(*ssm.Options)) (*ssm.GetParameterOutput, error)
}

type VaultStore struct {
	svc   paramStoreApi
	ctx   context.Context
	token string
}

func NewVaultStore(ctx context.Context) (*VaultStore, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		log.Errorf("unable to load SDK config, %v", err)
		return nil, err
	}
	c := ssm.NewFromConfig(cfg)

	return &VaultStore{
		svc: c,
		ctx: ctx,
	}, nil

}

func (imp *VaultStore) getTokenValue(v *retrieveStrategy) (string, error) {
	config := vault.DefaultConfig()

	config.Address = "http://127.0.0.1:8200"

	client, err := vault.NewClient(config)
	if err != nil {
		log.Errorf("unable to initialize Vault client: %v", err)
	}

	// Authenticate
	client.SetToken("dev-only-token")

	secretData := map[string]interface{}{
		"password": "Hashi123",
	}

	// Write a secret
	_, err = client.KVv2("secret").Put(context.Background(), "my-secret-password", secretData)
	if err != nil {
		log.Errorf("unable to write secret: %v", err)
	}

	fmt.Println("Secret written successfully.")

	// Read a secret from the default mount path for KV v2 in dev mode, "secret"
	secret, err := client.KVv2("secret").Get(context.Background(), "my-secret-password")
	if err != nil {
		log.Errorf("unable to read secret: %v", err)
	}

	value, ok := secret.Data["password"].(string)
	if !ok {
		log.Errorf("value type assertion failed: %T %#v", secret.Data["password"], secret.Data["password"])
	}

	if value != "Hashi123" {
		log.Errorf("unexpected password value %q retrieved from vault", value)
	}

	fmt.Println("Access granted!")
	return "", nil
}
