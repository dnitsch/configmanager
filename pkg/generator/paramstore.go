package generator

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/dnitsch/configmanager/pkg/log"
)

type paramStoreApi interface {
	GetParameter(ctx context.Context, params *ssm.GetParameterInput, optFns ...func(*ssm.Options)) (*ssm.GetParameterOutput, error)
}

type ParamStore struct {
	svc    paramStoreApi
	ctx    context.Context
	config *ParamStrConfig
	token  string
}

type ParamStrConfig struct {
	// reserved for potential future use
}

func NewParamStore(ctx context.Context) (*ParamStore, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		log.Errorf("unable to load SDK config, %v", err)
		return nil, err
	}
	c := ssm.NewFromConfig(cfg)

	return &ParamStore{
		svc: c,
		ctx: ctx,
	}, nil
}

func (imp *ParamStore) setTokenVal(token string) {
	storeConf := &ParamStrConfig{}
	initialToken := ParseMetadata(token, storeConf)

	imp.config = storeConf
	imp.token = initialToken
}

func (imp *ParamStore) tokenVal(v *retrieveStrategy) (string, error) {
	log.Infof("%s", "Concrete implementation ParameterStore")
	log.Infof("ParamStore Token: %s", imp.token)

	input := &ssm.GetParameterInput{
		Name:           aws.String(v.stripPrefix(imp.token, ParamStorePrefix)),
		WithDecryption: aws.Bool(true),
	}
	ctx, cancel := context.WithCancel(imp.ctx)
	defer cancel()

	result, err := imp.svc.GetParameter(ctx, input)
	if err != nil {
		log.Errorf(implementationNetworkErr, ParamStorePrefix, err, imp.token)
		return "", err
	}

	if result.Parameter.Value != nil {
		return *result.Parameter.Value, nil
	}
	log.Errorf("value retrieved but empty for token: %v", imp.token)
	return "", nil
}
