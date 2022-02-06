package genvars

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	klog "k8s.io/klog/v2"
)

type SecretsMgr struct {
	svc   *secretsmanager.Client
	ctx   context.Context
	token string
}

func NewSecretsMgr(ctx context.Context) (*SecretsMgr, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		klog.Errorf("unable to load SDK config, %v", err)
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
	klog.Infof("%s", "Concrete implementation SecretsManager")
	klog.Infof("Getting Secret: %s", implmt.token)

	input := &secretsmanager.GetSecretValueInput{
		SecretId:     aws.String(stripPrefix(implmt.token, SecretMgrPrefix)),
		VersionStage: aws.String("AWSCURRENT"),
	}

	result, err := implmt.svc.GetSecretValue(implmt.ctx, input)
	if err != nil {
		klog.Errorf("SecretsMgr: %s", err)
		return "", err
	}

	return *result.SecretString, nil
}
