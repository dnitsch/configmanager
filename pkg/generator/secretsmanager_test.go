package generator

import (
	"context"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/dnitsch/configmanager/internal/testutils"
)

var (
	tsuccessSecret = "someVal"
)

type mockSecretsApi func(ctx context.Context, params *secretsmanager.GetSecretValueInput, optFns ...func(*secretsmanager.Options)) (*secretsmanager.GetSecretValueOutput, error)

func (m mockSecretsApi) GetSecretValue(ctx context.Context, params *secretsmanager.GetSecretValueInput, optFns ...func(*secretsmanager.Options)) (*secretsmanager.GetSecretValueOutput, error) {
	return m(ctx, params, optFns...)
}

func Test_GetSecretMgrVarHappy(t *testing.T) {
	tests := []struct {
		name       string
		token      string
		value      string
		mockClient func(t *testing.T) secretsMgrApi
		config     *GenVarsConfig
	}{
		{
			name:  "successVal",
			token: "AWSSECRETS#/token/1",
			value: tsuccessParam,
			mockClient: func(t *testing.T) secretsMgrApi {
				return mockSecretsApi(func(ctx context.Context, params *secretsmanager.GetSecretValueInput, optFns ...func(*secretsmanager.Options)) (*secretsmanager.GetSecretValueOutput, error) {
					t.Helper()
					if params.VersionStage == nil {
						t.Fatal("expect name to not be nil")
					}

					if strings.Contains(*params.SecretId, "#") {
						t.Errorf("incorrectly stripped token separator")
					}

					if strings.Contains(*params.SecretId, SecretMgrPrefix) {
						t.Errorf("incorrectly stripped prefix")
					}

					return &secretsmanager.GetSecretValueOutput{
						SecretString: &tsuccessSecret,
					}, nil
				})
			},
			config: NewConfig(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.config.WithTokenSeparator(tokenSeparator)
			rs := newRetrieveStrategy(NewDefatultStrategy(), *tt.config)

			rs.setImplementation(&SecretsMgr{svc: tt.mockClient(t), ctx: context.TODO()})
			rs.setToken(tt.token)
			want, err := rs.getTokenValue()
			if err != nil {
				t.Errorf("%v", err)
			}
			if want != tt.value {
				t.Errorf(testutils.TestPhrase, want, tt.value)
			}
		})
	}
}
