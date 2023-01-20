package generator

import (
	"context"
	"strings"
	"testing"

	gcpsecretspb "cloud.google.com/go/secretmanager/apiv1/secretmanagerpb"
	"github.com/dnitsch/configmanager/internal/testutils"
	"github.com/googleapis/gax-go/v2"
)

type mockGcpSecretsApi func(ctx context.Context, req *gcpsecretspb.AccessSecretVersionRequest, opts ...gax.CallOption) (*gcpsecretspb.AccessSecretVersionResponse, error)

func (m mockGcpSecretsApi) AccessSecretVersion(ctx context.Context, req *gcpsecretspb.AccessSecretVersionRequest, opts ...gax.CallOption) (*gcpsecretspb.AccessSecretVersionResponse, error) {
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
				return mockGcpSecretsApi(func(ctx context.Context, req *gcpsecretspb.AccessSecretVersionRequest, opts ...gax.CallOption) (*gcpsecretspb.AccessSecretVersionResponse, error) {
					t.Helper()
					if req.Name == "" {
						t.Fatal("expect name to not be nil")
					}

					if strings.Contains(req.Name, "#") {
						t.Errorf("incorrectly stripped token separator")
					}

					if strings.Contains(req.Name, string(GcpSecretsPrefix)) {
						t.Errorf("incorrectly stripped prefix")
					}

					return &gcpsecretspb.AccessSecretVersionResponse{
						Payload: &gcpsecretspb.SecretPayload{Data: []byte(tsuccessSecret)},
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

			rs.setImplementation(&GcpSecrets{svc: tt.mockClient(t), ctx: context.TODO(), close: func() error { return nil }})
			rs.setToken(tt.token)
			got, err := rs.getTokenValue()
			if err != nil {
				t.Errorf("%v", err)
			}
			if got != tt.value {
				t.Errorf(testutils.TestPhrase, got, tt.value)
			}
		})
	}
}
