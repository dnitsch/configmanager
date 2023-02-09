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
	svc    secretsMgrApi
	ctx    context.Context
	config TokenConfigVars
	token  string
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
		ctx: ctx,
	}, nil

}

// func(imp *SecretsMgr) getTokenConfig() AdditionalVars {
// 	return
// }

func (imp *SecretsMgr) setToken(token string) {
	ct := (GenVarsConfig{}).ParseTokenVars(token)
	imp.config = ct
	imp.token = ct.Token
}

func (imp *SecretsMgr) getTokenValue(v *retrieveStrategy) (string, error) {

	log.Infof("%s", "Concrete implementation SecretsManager")

	version := "AWSCURRENT"
	if imp.config.Version != "" {
		version = imp.config.Version
	}

	log.Infof("Getting Secret: %s @version: %s", imp.token, version)

	input := &secretsmanager.GetSecretValueInput{
		SecretId:     aws.String(v.stripPrefix(imp.token, SecretMgrPrefix)),
		VersionStage: aws.String(version),
	}

	ctx, cancel := context.WithCancel(imp.ctx)
	defer cancel()

	result, err := imp.svc.GetSecretValue(ctx, input)
	if err != nil {
		log.Errorf(implementationNetworkErr, SecretMgrPrefix, err, imp.token)
		return "", err
	}

	if result.SecretString != nil {
		return *result.SecretString, nil
	}

	if len(result.SecretBinary) > 0 {
		return string(result.SecretBinary), nil
	}

	log.Errorf("value retrieved but empty for token: %v", imp.token)
	return "", nil
}
