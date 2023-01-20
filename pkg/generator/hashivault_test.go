package generator

import (
	"context"
	"fmt"
	"testing"

	"github.com/dnitsch/configmanager/internal/testutils"
	vault "github.com/hashicorp/vault/api"
)

func TestMountPathExtract(t *testing.T) {
	ttests := map[string]struct {
		token          string
		tokenSeparator string
		keySeparator   string
		expect         string
	}{
		"without leading slash":               {"VAULT://secret/demo/configmanager", "://", "|", "secret"},
		"with leading slash":                  {"VAULT:///secret/demo/configmanager", "://", "|", "secret"},
		"with underscore in path name":        {"VAULT://_secret/demo/configmanager", "://", "|", "_secret"},
		"with double underscore in path name": {"VAULT://__secret/demo/configmanager", "://", "|", "__secret"},
	}
	for name, tt := range ttests {
		t.Run(name, func(t *testing.T) {
			strippedToken := stripPrefix(tt.token, HashicorpVaultPrefix, tt.tokenSeparator, tt.keySeparator)
			got := splitToken(strippedToken)
			if got.path != tt.expect {
				t.Errorf("got %q, expected %q", got, tt.expect)
			}
		})
	}
}

type mockVaultApi func(ctx context.Context, secretPath string) (*vault.KVSecret, error)

func (m mockVaultApi) Get(ctx context.Context, secretPath string) (*vault.KVSecret, error) {
	return m(ctx, secretPath)
}

func TestVaultScenarios(t *testing.T) {
	ttests := map[string]struct {
		token       string
		conf        GenVarsConfig
		expect      string
		vaultClient func(t *testing.T) hashiVaultApi
	}{
		"happy return": {"VAULT://secret/foo", GenVarsConfig{tokenSeparator: "://", keySeparator: "|"}, `{"foo":"test2130-9sd-0ds"}`,
			func(t *testing.T) hashiVaultApi {
				return mockVaultApi(func(ctx context.Context, secretPath string) (*vault.KVSecret, error) {
					t.Helper()
					if secretPath != "secret/foo" {
						t.Errorf("got %v; want %s", secretPath, `secret/foo`)
					}
					m := make(map[string]interface{})
					m["foo"] = "test2130-9sd-0ds"
					return &vault.KVSecret{Data: m}, nil
				})
			},
		},
		"incorrect json": {"VAULT://secret/foo", GenVarsConfig{tokenSeparator: "://", keySeparator: "|"}, `json: unsupported type: func() error`,
			func(t *testing.T) hashiVaultApi {
				return mockVaultApi(func(ctx context.Context, secretPath string) (*vault.KVSecret, error) {
					t.Helper()
					if secretPath != "secret/foo" {
						t.Errorf("got %v; want %s", secretPath, `secret/foo`)
					}
					m := make(map[string]interface{})
					m["error"] = func() error { return fmt.Errorf("ddodod") }
					return &vault.KVSecret{Data: m}, nil
				})
			},
		},
		"another return": {
			token:  "VAULT://secret/some/other/foo2",
			conf:   GenVarsConfig{tokenSeparator: "://", keySeparator: "|"},
			expect: `{"foo1":"test2130-9sd-0ds","foo2":"dsfsdf3454456"}`,
			vaultClient: func(t *testing.T) hashiVaultApi {
				return mockVaultApi(func(ctx context.Context, secretPath string) (*vault.KVSecret, error) {
					t.Helper()
					if secretPath != "secret/some/other/foo2" {
						t.Errorf("got %v; want %s", secretPath, `secret/some/other/foo2`)
					}
					m := make(map[string]interface{})
					m["foo1"] = "test2130-9sd-0ds"
					m["foo2"] = "dsfsdf3454456"
					return &vault.KVSecret{Data: m}, nil
				})
			},
		},
		"not found": {"VAULT://secret/foo", GenVarsConfig{tokenSeparator: "://", keySeparator: "|"}, `secret not found`,
			func(t *testing.T) hashiVaultApi {
				return mockVaultApi(func(ctx context.Context, secretPath string) (*vault.KVSecret, error) {
					t.Helper()
					if secretPath != "secret/foo" {
						t.Errorf("got %v; want %s", secretPath, `secret/foo`)
					}
					return nil, fmt.Errorf("secret not found")
				})
			},
		},
		"403": {
			token:  "VAULT://secret/some/other/foo2",
			conf:   GenVarsConfig{tokenSeparator: "://", keySeparator: "|"},
			expect: `client 403`,
			vaultClient: func(t *testing.T) hashiVaultApi {
				return mockVaultApi(func(ctx context.Context, secretPath string) (*vault.KVSecret, error) {
					t.Helper()
					if secretPath != "secret/some/other/foo2" {
						t.Errorf("got %v; want %s", secretPath, `secret/some/other/foo2`)
					}
					return nil, fmt.Errorf("client 403")
				})
			},
		},
		"found but empty": {
			token:  "VAULT://secret/some/other/foo2",
			conf:   GenVarsConfig{tokenSeparator: "://", keySeparator: "|"},
			expect: `{}`,
			vaultClient: func(t *testing.T) hashiVaultApi {
				return mockVaultApi(func(ctx context.Context, secretPath string) (*vault.KVSecret, error) {
					t.Helper()
					if secretPath != "secret/some/other/foo2" {
						t.Errorf("got %v; want %s", secretPath, `secret/some/other/foo2`)
					}
					m := make(map[string]interface{})
					return &vault.KVSecret{Data: m}, nil
				})
			},
		},
	}

	for name, tt := range ttests {
		t.Run(name, func(t *testing.T) {
			vs := &VaultStore{
				svc:   tt.vaultClient(t),
				ctx:   context.TODO(),
				token: tt.token,
			}

			rs := newRetrieveStrategy(NewDefatultStrategy(), tt.conf)
			rs.setImplementation(vs)
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
