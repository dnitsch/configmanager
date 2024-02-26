package generator

import (
	"context"
	"fmt"

	gcpsecrets "cloud.google.com/go/secretmanager/apiv1"
	gcpsecretspb "cloud.google.com/go/secretmanager/apiv1/secretmanagerpb"
	"github.com/dnitsch/configmanager/pkg/log"
	"github.com/googleapis/gax-go/v2"
)

type gcpSecretsApi interface {
	AccessSecretVersion(ctx context.Context, req *gcpsecretspb.AccessSecretVersionRequest, opts ...gax.CallOption) (*gcpsecretspb.AccessSecretVersionResponse, error)
}

type GcpSecrets struct {
	svc    gcpSecretsApi
	ctx    context.Context
	config TokenConfigVars
	close  func() error
	token  string
}

func NewGcpSecrets(ctx context.Context) (*GcpSecrets, error) {

	c, err := gcpsecrets.NewClient(ctx)
	if err != nil {
		return nil, err
	}
	// defer c.Close()

	return &GcpSecrets{
		svc:   c,
		ctx:   ctx,
		close: c.Close,
	}, nil
}

func (imp *GcpSecrets) setTokenVal(token string) {
	ct := (GenVarsConfig{}).ParseTokenVars(token)
	imp.config = ct
	imp.token = ct.Token
}

func (imp *GcpSecrets) tokenVal(v *retrieveStrategy) (string, error) {
	defer imp.close()
	log.Infof("%s", "Concrete implementation GcpSecrets")

	version := "latest"
	if imp.config.Version != "" {
		version = imp.config.Version
	}

	log.Infof("Getting Secret: %s @version: %s", imp.token, version)

	input := &gcpsecretspb.AccessSecretVersionRequest{
		Name: fmt.Sprintf("%s/versions/%s", v.stripPrefix(imp.token, GcpSecretsPrefix), version),
	}

	ctx, cancel := context.WithCancel(imp.ctx)
	defer cancel()

	result, err := imp.svc.AccessSecretVersion(ctx, input)

	if err != nil {
		log.Errorf(implementationNetworkErr, GcpSecretsPrefix, err, imp.token)
		return "", err
	}
	if result.Payload != nil {
		return string(result.GetPayload().GetData()), nil
	}

	log.Errorf("value retrieved but empty for token: %v", imp.token)
	return "", nil
}
