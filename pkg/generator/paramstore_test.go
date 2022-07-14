package generator

import (
	"context"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/aws/aws-sdk-go-v2/service/ssm/types"
	"github.com/dnitsch/configmanager/internal/testutils"
)

var (
	tsuccessParam                   = "someVal"
	tsuccessObj   map[string]string = map[string]string{"AWSPARAMSTR#/token/1": "someVal"}
)

type mockParamApi func(ctx context.Context, params *ssm.GetParameterInput, optFns ...func(*ssm.Options)) (*ssm.GetParameterOutput, error)

func (m mockParamApi) GetParameter(ctx context.Context, params *ssm.GetParameterInput, optFns ...func(*ssm.Options)) (*ssm.GetParameterOutput, error) {
	return m(ctx, params, optFns...)
}

func Test_GetParamStoreVarHappy(t *testing.T) {
	tests := []struct {
		name       string
		token      string
		value      string
		mockClient func(t *testing.T) paramStoreApi
		genVars    *GenVars
	}{
		{
			name:  "successVal",
			token: "AWSPARAMSTR#/token/1",
			value: tsuccessParam,
			mockClient: func(t *testing.T) paramStoreApi {
				return mockParamApi(func(ctx context.Context, params *ssm.GetParameterInput, optFns ...func(*ssm.Options)) (*ssm.GetParameterOutput, error) {
					t.Helper()
					if params.Name == nil {
						t.Fatal("expect name to not be nil")
					}

					if strings.Contains(*params.Name, "#") {
						t.Errorf("incorrectly stripped token separator")
					}

					if strings.Contains(*params.Name, paramStorePrefix) {
						t.Errorf("incorrectly stripped prefix")
					}

					if !params.WithDecryption {
						t.Fatal("expect WithDecryption to not be false")
					}

					return &ssm.GetParameterOutput{
						Parameter: &types.Parameter{Value: &tsuccessParam},
					}, nil
				})
			},
			genVars: &GenVars{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.genVars.config = GenVarsConfig{tokenSeparator: tokenSeparator}
			tt.genVars.setImplementation(&ParamStore{svc: tt.mockClient(t)})
			tt.genVars.setToken(tt.token)
			want, err := tt.genVars.getTokenValue()
			if err != nil {
				t.Errorf("%v", err)
			}
			if want != tt.value {
				t.Errorf(testutils.TestPhrase, want, tt.value)
			}
		})
	}
}
