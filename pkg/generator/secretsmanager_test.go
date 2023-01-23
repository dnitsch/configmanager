package generator

import (
	"context"
	"fmt"
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

func awsSecretsMgrGetChecker(t *testing.T, params *secretsmanager.GetSecretValueInput) {
	if params.VersionStage == nil {
		t.Fatal("expect name to not be nil")
	}

	if strings.Contains(*params.SecretId, "#") {
		t.Errorf("incorrectly stripped token separator")
	}

	if strings.Contains(*params.SecretId, string(SecretMgrPrefix)) {
		t.Errorf("incorrectly stripped prefix")
	}
}

func Test_GetSecretMgr(t *testing.T) {
	tests := map[string]struct {
		token          string
		keySeparator   string
		tokenSeparator string
		expect         string
		mockClient     func(t *testing.T) secretsMgrApi
		config         *GenVarsConfig
	}{
		"success": {"AWSSECRETS#/token/1", "|", "#", tsuccessParam, func(t *testing.T) secretsMgrApi {
			return mockSecretsApi(func(ctx context.Context, params *secretsmanager.GetSecretValueInput, optFns ...func(*secretsmanager.Options)) (*secretsmanager.GetSecretValueOutput, error) {
				t.Helper()
				awsSecretsMgrGetChecker(t, params)
				return &secretsmanager.GetSecretValueOutput{
					SecretString: &tsuccessSecret,
				}, nil
			})
		}, NewConfig(),
		},
		"success with binary": {"AWSSECRETS#/token/1", "|", "#", tsuccessParam, func(t *testing.T) secretsMgrApi {
			return mockSecretsApi(func(ctx context.Context, params *secretsmanager.GetSecretValueInput, optFns ...func(*secretsmanager.Options)) (*secretsmanager.GetSecretValueOutput, error) {
				t.Helper()
				awsSecretsMgrGetChecker(t, params)
				return &secretsmanager.GetSecretValueOutput{
					SecretBinary: []byte(tsuccessParam),
				}, nil
			})
		}, NewConfig(),
		},
		"errored": {"AWSSECRETS#/token/1", "|", "#", "unable to retrieve secret", func(t *testing.T) secretsMgrApi {
			return mockSecretsApi(func(ctx context.Context, params *secretsmanager.GetSecretValueInput, optFns ...func(*secretsmanager.Options)) (*secretsmanager.GetSecretValueOutput, error) {
				t.Helper()
				awsSecretsMgrGetChecker(t, params)
				return nil, fmt.Errorf("unable to retrieve secret")
			})
		}, NewConfig(),
		},
		"ok but empty": {"AWSSECRETS#/token/1", "|", "#", "", func(t *testing.T) secretsMgrApi {
			return mockSecretsApi(func(ctx context.Context, params *secretsmanager.GetSecretValueInput, optFns ...func(*secretsmanager.Options)) (*secretsmanager.GetSecretValueOutput, error) {
				t.Helper()
				awsSecretsMgrGetChecker(t, params)
				return &secretsmanager.GetSecretValueOutput{
					SecretString: nil,
				}, nil
			})
		}, NewConfig(),
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			tt.config.WithTokenSeparator(tt.tokenSeparator).WithKeySeparator(tt.keySeparator)
			impl, _ := NewSecretsMgr(context.TODO())
			impl.svc = tt.mockClient(t)
			rs := newRetrieveStrategy(NewDefatultStrategy(), *tt.config)

			rs.setImplementation(impl)
			rs.setToken(tt.token)
			got, err := rs.getTokenValue()
			if err != nil {
				if err.Error() != tt.expect {
					t.Errorf(testutils.TestPhrase, err.Error(), tt.expect)
				}
				return
			}
			if got != tt.expect {
				t.Errorf(testutils.TestPhrase, got, tt.expect)
			}
		})
	}
}
