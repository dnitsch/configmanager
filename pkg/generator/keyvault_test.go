package generator

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/keyvault/azsecrets"
	"github.com/dnitsch/configmanager/internal/testutils"
)

func Test_azSplitToken(t *testing.T) {
	tests := []struct {
		name   string
		token  string
		expect azVaultHelper
	}{
		{
			name:  "simple_with_preceding_slash",
			token: "/test-vault/somejsontest",
			expect: azVaultHelper{
				vaultUri: "https://test-vault.vault.azure.net",
				token:    "somejsontest",
			},
		},
		{
			name:  "missing_initial_slash",
			token: "test-vault/somejsontest",
			expect: azVaultHelper{
				vaultUri: "https://test-vault.vault.azure.net",
				token:    "somejsontest",
			},
		},
		{
			name:  "missing_initial_slash_multislash_secretname",
			token: "test-vault/some/json/test",
			expect: azVaultHelper{
				vaultUri: "https://test-vault.vault.azure.net",
				token:    "some/json/test",
			},
		},
		{
			name:  "with_initial_slash_multislash_secretname",
			token: "test-vault//some/json/test",
			expect: azVaultHelper{
				vaultUri: "https://test-vault.vault.azure.net",
				token:    "/some/json/test",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := azSplitToken(tt.token)
			if got.token != tt.expect.token {
				t.Errorf(testutils.TestPhrase, tt.expect.token, got.token)
			}
			if got.vaultUri != tt.expect.vaultUri {
				t.Errorf(testutils.TestPhrase, tt.expect.vaultUri, got.vaultUri)
			}
		})
	}
}

var (
	tazkvsuccessObj map[string]string = map[string]string{fmt.Sprintf("%s#/token/1", AzKeyVaultSecretsPrefix): tsuccessParam}
)

type mockAzKvSecretApi func(ctx context.Context, name string, version string, options *azsecrets.GetSecretOptions) (azsecrets.GetSecretResponse, error)

func (m mockAzKvSecretApi) GetSecret(ctx context.Context, name string, version string, options *azsecrets.GetSecretOptions) (azsecrets.GetSecretResponse, error) {
	return m(ctx, name, version, options)
}

func Test_GetAzKeyVaultSecretVarHappy(t *testing.T) {

	tests := []struct {
		name       string
		token      string
		value      string
		mockClient func(t *testing.T) kvApi
		config     *GenVarsConfig
	}{
		{
			name:  "successVal",
			token: "AZKVSECRET#/test-vault//token/1",
			value: tsuccessParam,
			mockClient: func(t *testing.T) kvApi {
				return mockAzKvSecretApi(func(ctx context.Context, name string, version string, options *azsecrets.GetSecretOptions) (azsecrets.GetSecretResponse, error) {
					t.Helper()
					if name == "" {
						t.Errorf("expect name to not be nil")
					}
					if name != "/token/1" {
						t.Errorf(testutils.TestPhrase, "/token/1", name)
					}

					if strings.Contains(name, "#") {
						t.Errorf("incorrectly stripped token separator")
					}

					if strings.Contains(name, AzKeyVaultSecretsPrefix) {
						t.Errorf("incorrectly stripped prefix")
					}

					if version != "" {
						t.Fatal("expect version to be \"\" an empty string ")
					}

					resp := azsecrets.GetSecretResponse{}
					resp.Value = &tsuccessParam
					return resp, nil
				})
			},
			config: NewConfig(),
		},
		{
			name:  "successVal with keyseparator",
			token: "AZKVSECRET#/test-vault/token/1|somekey",
			value: tsuccessParam,
			mockClient: func(t *testing.T) kvApi {
				return mockAzKvSecretApi(func(ctx context.Context, name string, version string, options *azsecrets.GetSecretOptions) (azsecrets.GetSecretResponse, error) {
					t.Helper()
					if name == "" {
						t.Error("expect name to not be nil")
					}
					if name != "token/1" {
						t.Errorf(testutils.TestPhrase, "token/1", name)
					}
					if strings.Contains(name, "#") {
						t.Errorf("incorrectly stripped token separator")
					}

					if strings.Contains(name, AzKeyVaultSecretsPrefix) {
						t.Errorf("incorrectly stripped prefix")
					}

					if version != "" {
						t.Fatal("expect version to be \"\" an empty string ")
					}
					resp := azsecrets.GetSecretResponse{}
					resp.Value = &tsuccessParam
					return resp, nil
				})
			},
			config: NewConfig(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			kvStr, err := NewKvScrtStoreWithToken(context.TODO(), tt.token, "#", "|")
			if err != nil {
				t.Errorf("failed to init azkvstore")
			}
			kvStr.svc = tt.mockClient(t)
			rs := newRetrieveStrategy(NewDefatultStrategy(), *tt.config)
			rs.setImplementation(kvStr)
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
