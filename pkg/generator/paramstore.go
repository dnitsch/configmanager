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
	svc   paramStoreApi
	token string
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
	}, nil

}

func (paramStr *ParamStore) setToken(token string) {
	paramStr.token = token
}

func (implmt *ParamStore) setValue(val string) {
}

func (imp *ParamStore) getTokenValue(v *GenVars) (string, error) {
	log.Infof("%s", "Concrete implementation ParameterStore SecureString")
	log.Infof("ParamStore Token: %s", imp.token)

	input := &ssm.GetParameterInput{
		Name:           aws.String(v.stripPrefix(imp.token, ParamStorePrefix)),
		WithDecryption: true,
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	result, err := imp.svc.GetParameter(ctx, input)
	if err != nil {
		log.Errorf("ParamStore: %s", err)
		return "", err
	}

	return *result.Parameter.Value, nil
}
