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
		genVars    *genVars
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

					if strings.Contains(*params.SecretId, paramStorePrefix) {
						t.Errorf("incorrectly stripped prefix")
					}

					return &secretsmanager.GetSecretValueOutput{
						SecretString: &tsuccessSecret,
					}, nil
				})
			},
			genVars: &genVars{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.genVars.config = GenVarsConfig{tokenSeparator: tokenSeparator}
			tt.genVars.setImplementation(&SecretsMgr{svc: tt.mockClient(t)})
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
