package genvars

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	klog "k8s.io/klog/v2"
)

type ParamStore struct {
	svc   *ssm.Client
	ctx   context.Context
	token string
}

func NewParamStore(ctx context.Context) (*ParamStore, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		klog.Errorf("unable to load SDK config, %v", err)
		return nil, err
	}
	initService := ssm.NewFromConfig(cfg)

	return &ParamStore{
		svc: initService,
		ctx: ctx,
	}, nil

}

func (paramStr *ParamStore) setToken(token string) {
	paramStr.token = token
}

func (imp *ParamStore) getTokenValue(v *genVars) (string, error) {
	klog.Infof("%s", "Concrete implementation ParameterStore SecureString")
	klog.Infof("ParamStore Token: %s", imp.token)

	input := &ssm.GetParameterInput{
		Name:           aws.String(stripPrefix(imp.token, ParamStorePrefix)),
		WithDecryption: true,
	}

	result, err := imp.svc.GetParameter(imp.ctx, input)
	if err != nil {
		klog.Errorf("ParamStore: %s", err)
		return "", err
	}

	return *result.Parameter.Value, nil
}
