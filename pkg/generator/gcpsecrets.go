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
	svc   gcpSecretsApi
	ctx   context.Context
	close func() error
	token string
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

func (imp *GcpSecrets) setToken(token string) {
	imp.token = token
}

func (imp *GcpSecrets) setValue(val string) {
}

func (imp *GcpSecrets) getTokenValue(v *retrieveStrategy) (string, error) {
	defer imp.close()

	log.Infof("%s", "Concrete implementation GcpSecrets")
	log.Infof("Getting Secret: %s", imp.token)

	input := &gcpsecretspb.AccessSecretVersionRequest{
		Name: fmt.Sprintf("%s/versions/latest", v.stripPrefix(imp.token, GcpSecretsPrefix)),
	}
	ctx, cancel := context.WithCancel(imp.ctx)
	defer cancel()

	result, err := imp.svc.AccessSecretVersion(ctx, input)

	if err != nil {
		log.Errorf(implementationNetworkErr, GcpSecretsPrefix, err, imp.token)
		return "", err
	}
	if result.Payload != nil {
		return string(result.Payload.Data), nil
	}

	log.Errorf("value retrieved but empty for token: %v", imp.token)
	return "", nil
}
