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

func (implmt *SecretsMgr) setToken(token string) {
	implmt.token = token
}

func (implmt *SecretsMgr) setValue(val string) {
}

func (implmt *SecretsMgr) getTokenValue(v *GenVars) (string, error) {
	log.Infof("%s", "Concrete implementation SecretsManager")
	log.Infof("Getting Secret: %s", implmt.token)

	input := &secretsmanager.GetSecretValueInput{
		SecretId:     aws.String(v.stripPrefix(implmt.token, SecretMgrPrefix)),
		VersionStage: aws.String("AWSCURRENT"),
	}

	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()

	result, err := implmt.svc.GetSecretValue(ctx, input)
	if err != nil {
		log.Errorf("SecretsMgr: %s", err)
		return "", err
	}

	return *result.SecretString, nil
}
