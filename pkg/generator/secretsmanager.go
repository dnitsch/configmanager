package generator

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/dnitsch/configmanager/pkg/log"
)

type secretsMgrApi interface {
	GetSecretValue(ctx context.Context, params *secretsmanager.GetSecretValueInput, optFns ...func(*secretsmanager.Options)) (*secretsmanager.GetSecretValueOutput, error)
}

type SecretsMgr struct {
	svc   secretsMgrApi
	token string
}

func NewSecretsMgr(ctx context.Context) (*SecretsMgr, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		log.Errorf("unable to load SDK config, %v", err)
		return nil, err
	}
	c := secretsmanager.NewFromConfig(cfg)

	return &SecretsMgr{
		svc: c,
	}, nil

}

func (imp *SecretsMgr) setToken(token string) {
	imp.token = token
}

func (imp *SecretsMgr) setValue(val string) {
}

func (imp *SecretsMgr) getTokenValue(v *GenVars) (string, error) {
	log.Infof("%s", "Concrete implementation SecretsManager")
	log.Infof("Getting Secret: %s", imp.token)

	input := &secretsmanager.GetSecretValueInput{
		SecretId:     aws.String(v.stripPrefix(imp.token, SecretMgrPrefix)),
		VersionStage: aws.String("AWSCURRENT"),
	}

	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()

	result, err := imp.svc.GetSecretValue(ctx, input)
	if err != nil {
		log.Errorf("SecretsMgr: %s", err)
		return "", err
	}
	if result.SecretString != nil {
		return *result.SecretString, nil
	}
	log.Errorf("value retrieved but empty for token: %v", imp.token)
	return "", nil
}
