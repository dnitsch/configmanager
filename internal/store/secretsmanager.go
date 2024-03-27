package store

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsConf "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/dnitsch/configmanager/internal/config"
	"github.com/dnitsch/configmanager/pkg/log"
)

type secretsMgrApi interface {
	GetSecretValue(ctx context.Context, params *secretsmanager.GetSecretValueInput, optFns ...func(*secretsmanager.Options)) (*secretsmanager.GetSecretValueOutput, error)
}

type SecretsMgr struct {
	svc    secretsMgrApi
	ctx    context.Context
	config *SecretsMgrConfig
	token  *config.ParsedTokenConfig
}

type SecretsMgrConfig struct {
	Version string `json:"version"`
}

func NewSecretsMgr(ctx context.Context) (*SecretsMgr, error) {
	cfg, err := awsConf.LoadDefaultConfig(ctx)
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

func (imp *SecretsMgr) SetToken(token *config.ParsedTokenConfig) {
	storeConf := &SecretsMgrConfig{}
	token.ParseMetadata(storeConf)
	imp.token = token
	imp.config = storeConf
}

func (imp *SecretsMgr) Token() (string, error) {
	log.Infof("Concrete implementation SecretsManager")
	log.Infof("SecretsManager Token: %s", imp.token.String())

	version := "AWSCURRENT"
	if imp.config.Version != "" {
		version = imp.config.Version
	}

	log.Infof("Getting Secret: %s @version: %s", imp.token, version)

	input := &secretsmanager.GetSecretValueInput{
		SecretId:     aws.String(imp.token.StoreToken()),
		VersionStage: aws.String(version),
	}

	ctx, cancel := context.WithCancel(imp.ctx)
	defer cancel()

	result, err := imp.svc.GetSecretValue(ctx, input)
	if err != nil {
		log.Errorf(implementationNetworkErr, imp.token.Prefix(), err, imp.token.String())
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
