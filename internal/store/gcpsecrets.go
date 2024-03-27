package store

import (
	"context"
	"fmt"

	gcpsecrets "cloud.google.com/go/secretmanager/apiv1"
	gcpsecretspb "cloud.google.com/go/secretmanager/apiv1/secretmanagerpb"
	"github.com/dnitsch/configmanager/internal/config"
	"github.com/dnitsch/configmanager/pkg/log"
	"github.com/googleapis/gax-go/v2"
)

type gcpSecretsApi interface {
	AccessSecretVersion(ctx context.Context, req *gcpsecretspb.AccessSecretVersionRequest, opts ...gax.CallOption) (*gcpsecretspb.AccessSecretVersionResponse, error)
}

type GcpSecrets struct {
	svc    gcpSecretsApi
	ctx    context.Context
	config *GcpSecretsConfig
	close  func() error
	token  *config.ParsedTokenConfig
}

type GcpSecretsConfig struct {
	Version string `json:"version"`
}

func NewGcpSecrets(ctx context.Context) (*GcpSecrets, error) {

	c, err := gcpsecrets.NewClient(ctx)
	if err != nil {
		return nil, err
	}
	return &GcpSecrets{
		svc:   c,
		ctx:   ctx,
		close: c.Close,
	}, nil
}

func (imp *GcpSecrets) SetToken(token *config.ParsedTokenConfig) {
	storeConf := &GcpSecretsConfig{}
	token.ParseMetadata(storeConf)
	imp.token = token
	imp.config = storeConf
}

func (imp *GcpSecrets) Token() (string, error) {
	// Close client currently as new one would be created per iteration
	defer imp.close()

	log.Info("Concrete implementation GcpSecrets")
	log.Infof("GcpSecrets Token: %s", imp.token.String())

	version := "latest"
	if imp.config.Version != "" {
		version = imp.config.Version
	}

	log.Infof("Getting Secret: %s @version: %s", imp.token, version)

	input := &gcpsecretspb.AccessSecretVersionRequest{
		Name: fmt.Sprintf("%s/versions/%s", imp.token.StoreToken(), version),
	}

	ctx, cancel := context.WithCancel(imp.ctx)
	defer cancel()

	result, err := imp.svc.AccessSecretVersion(ctx, input)

	if err != nil {
		log.Errorf(implementationNetworkErr, imp.token.Prefix(), err, imp.token.String())
		return "", err
	}
	if result.Payload != nil {
		return string(result.Payload.Data), nil
	}

	log.Errorf("value retrieved but empty for token: %v", imp.token)
	return "", nil
}
