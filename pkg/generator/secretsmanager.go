package generator

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/dnitsch/genvars/pkg/log"
)

type SecretsMgr struct {
	svc   *secretsmanager.Client
	ctx   context.Context
	token string
}

func NewSecretsMgr(ctx context.Context) (*SecretsMgr, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		log.Errorf("unable to load SDK config, %v", err)
		return nil, err
	}
	initService := secretsmanager.NewFromConfig(cfg)

	return &SecretsMgr{
		svc: initService,
		ctx: ctx,
	}, nil

}

func (implmt *SecretsMgr) setToken(token string) {
	implmt.token = token
}

func (implmt *SecretsMgr) getTokenValue(v *genVars) (string, error) {
	log.Infof("%s", "Concrete implementation SecretsManager")
	log.Infof("Getting Secret: %s", implmt.token)

	input := &secretsmanager.GetSecretValueInput{
		SecretId:     aws.String(v.stripPrefix(implmt.token, SecretMgrPrefix)),
		VersionStage: aws.String("AWSCURRENT"),
	}

	result, err := implmt.svc.GetSecretValue(implmt.ctx, input)
	if err != nil {
		log.Errorf("SecretsMgr: %s", err)
		return "", err
	}

	return *result.SecretString, nil
}
