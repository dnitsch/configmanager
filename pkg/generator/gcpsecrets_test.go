package generator

import (
	"context"
	"strings"
	"testing"

	gcpsecretspb "cloud.google.com/go/secretmanager/apiv1/secretmanagerpb"
	"github.com/dnitsch/configmanager/internal/testutils"
	"github.com/googleapis/gax-go/v2"
)

var (
	tsuccessSecret = "someVal"
)

type mockSecretsApi func(ctx context.Context, params *gcpsecretspb.AccessSecretVersionRequest, opts ...gax.CallOption) (*gcpsecretspb.AccessSecretVersionResponse, error)

func (m mockSecretsApi) AccessSecretVersion(ctx context.Context, req *gcpsecretspb.AccessSecretVersionRequest, opts ...gax.CallOption) (*gcpsecretspb.AccessSecretVersionResponse, error) {
	return m(ctx, req, opts...)
}

func Test_GetGcpSecretVarHappy(t *testing.T) {
	tests := []struct {
		name       string
		token      string
		value      string
		mockClient func(t *testing.T) gcpSecretsApi
		config     *GenVarsConfig
	}{
		{
			name:  "successVal",
			token: "GCPSECRETS#/token/1",
			value: tsuccessParam,
			mockClient: func(t *testing.T) gcpSecretsApi {
				return mockSecretsApi(func(ctx context.Context, params *gcpsecretspb.AccessSecretVersionRequest, opts ...gax.CallOption) (*gcpsecretspb.AccessSecretVersionResponse, error) {
					t.Helper()
					if params.Name == nil {
						t.Fatal("expect name to not be nil")
					}

					if strings.Contains(*params.Name, "#") {
						t.Errorf("incorrectly stripped token separator")
					}

					if strings.Contains(*params.Name, SecretMgrPrefix) {
						t.Errorf("incorrectly stripped prefix")
					}

					return &gcpsecretspb.AccessSecretVersionResponse{
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
