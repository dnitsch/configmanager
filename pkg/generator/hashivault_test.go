package generator

import (
	"context"
	"fmt"
	"os"
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
		"without leading slash":               {"VAULT://secret___/demo/configmanager", "://", "|", "secret"},
		"with leading slash":                  {"VAULT:///secret___/demo/configmanager", "://", "|", "secret"},
		"with underscore in path name":        {"VAULT://_secret___/demo/configmanager", "://", "|", "_secret"},
		"with double underscore in path name": {"VAULT://__secret___/demo/configmanager", "://", "|", "__secret"},
		"with multiple paths in mountpath":    {"VAULT://secret/bar/path___/demo/configmanager", "://", "|", "secret/bar/path"},
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
		token      string
		conf       GenVarsConfig
		expect     string
		mockClient func(t *testing.T) hashiVaultApi
		setupEnv   func() func()
	}{
		"happy return": {"VAULT://secret___/foo", GenVarsConfig{tokenSeparator: "://", keySeparator: "|"}, `{"foo":"test2130-9sd-0ds"}`,
			func(t *testing.T) hashiVaultApi {
				return mockVaultApi(func(ctx context.Context, secretPath string) (*vault.KVSecret, error) {
					t.Helper()
					if secretPath != "foo" {
						t.Errorf("got %v; want %s", secretPath, `foo`)
					}
					m := make(map[string]interface{})
					m["foo"] = "test2130-9sd-0ds"
					return &vault.KVSecret{Data: m}, nil
				})
			},
			func() func() {
				os.Setenv("VAULT_TOKEN", "129378y1231283")
				return func() {
					os.Clearenv()
				}
			},
		},
		"incorrect json": {"VAULT://secret___/foo", GenVarsConfig{tokenSeparator: "://", keySeparator: "|"}, `json: unsupported type: func() error`,
			func(t *testing.T) hashiVaultApi {
				return mockVaultApi(func(ctx context.Context, secretPath string) (*vault.KVSecret, error) {
					t.Helper()
					if secretPath != "foo" {
						t.Errorf("got %v; want %s", secretPath, `foo`)
					}
					m := make(map[string]interface{})
					m["error"] = func() error { return fmt.Errorf("ddodod") }
					return &vault.KVSecret{Data: m}, nil
				})
			},
			func() func() {
				os.Setenv("VAULT_TOKEN", "129378y1231283")
				return func() {
					os.Clearenv()
				}
			},
		},
		"another return": {
			"VAULT://secret/engine1___/some/other/foo2",
			GenVarsConfig{tokenSeparator: "://", keySeparator: "|"},
			`{"foo1":"test2130-9sd-0ds","foo2":"dsfsdf3454456"}`,
			func(t *testing.T) hashiVaultApi {
				return mockVaultApi(func(ctx context.Context, secretPath string) (*vault.KVSecret, error) {
					t.Helper()
					if secretPath != "some/other/foo2" {
						t.Errorf("got %v; want %s", secretPath, `some/other/foo2`)
					}
					m := make(map[string]interface{})
					m["foo1"] = "test2130-9sd-0ds"
					m["foo2"] = "dsfsdf3454456"
					return &vault.KVSecret{Data: m}, nil
				})
			},
			func() func() {
				os.Setenv("VAULT_TOKEN", "129378y1231283")
				return func() {
					os.Clearenv()
				}
			},
		},
		"not found": {"VAULT://secret___/foo", GenVarsConfig{tokenSeparator: "://", keySeparator: "|"}, `secret not found`,
			func(t *testing.T) hashiVaultApi {
				return mockVaultApi(func(ctx context.Context, secretPath string) (*vault.KVSecret, error) {
					t.Helper()
					if secretPath != "foo" {
						t.Errorf("got %v; want %s", secretPath, `foo`)
					}
					return nil, fmt.Errorf("secret not found")
				})
			},
			func() func() {
				os.Setenv("VAULT_TOKEN", "129378y1231283")
				return func() {
					os.Clearenv()
				}
			},
		},
		"403": {"VAULT://secret___/some/other/foo2", GenVarsConfig{tokenSeparator: "://", keySeparator: "|"}, `client 403`,
			func(t *testing.T) hashiVaultApi {
				return mockVaultApi(func(ctx context.Context, secretPath string) (*vault.KVSecret, error) {
					t.Helper()
					if secretPath != "some/other/foo2" {
						t.Errorf("got %v; want %s", secretPath, `some/other/foo2`)
					}
					return nil, fmt.Errorf("client 403")
				})
			},
			func() func() {
				os.Setenv("VAULT_TOKEN", "129378y1231283")
				return func() {
					os.Clearenv()
				}
			},
		},
		"found but empty": {"VAULT://secret___/some/other/foo2", GenVarsConfig{tokenSeparator: "://", keySeparator: "|"}, `{}`, func(t *testing.T) hashiVaultApi {
			return mockVaultApi(func(ctx context.Context, secretPath string) (*vault.KVSecret, error) {
				t.Helper()
				if secretPath != "some/other/foo2" {
					t.Errorf("got %v; want %s", secretPath, `some/other/foo2`)
				}
				m := make(map[string]interface{})
				return &vault.KVSecret{Data: m}, nil
			})
		},
			func() func() {
				os.Setenv("VAULT_TOKEN", "129378y1231283")
				return func() {
					os.Clearenv()
				}
			},
		},
		"found but nil returned": {"VAULT://secret___/some/other/foo2", GenVarsConfig{tokenSeparator: "://", keySeparator: "|"}, "", func(t *testing.T) hashiVaultApi {
			return mockVaultApi(func(ctx context.Context, secretPath string) (*vault.KVSecret, error) {
				t.Helper()
				if secretPath != "some/other/foo2" {
					t.Errorf(testutils.TestPhrase, secretPath, `some/other/foo2`)
				}
				return &vault.KVSecret{Data: nil}, nil
			})
		},
			func() func() {
				os.Setenv("VAULT_TOKEN", "129378y1231283")
				return func() {
					os.Clearenv()
				}
			},
		},
		"vault rate limit incorrect": {"VAULT://secret___/some/other/foo2", GenVarsConfig{tokenSeparator: "://", keySeparator: "|"}, "unable to initialize Vault client: error encountered setting up default configuration: VAULT_RATE_LIMIT was provided but incorrectly formatted", func(t *testing.T) hashiVaultApi {
			return mockVaultApi(func(ctx context.Context, secretPath string) (*vault.KVSecret, error) {
				t.Helper()
				if secretPath != "some/other/foo2" {
					t.Errorf(testutils.TestPhrase, secretPath, `some/other/foo2`)
				}
				return &vault.KVSecret{Data: nil}, nil
			})
		},
			func() func() {
				os.Setenv("VAULT_TOKEN_INCORRECT", "")
				os.Setenv("VAULT_ADDR", "wrong://addr")
				os.Setenv("VAULT_RATE_LIMIT", "wrong")
				// os.Setenv("AWS_ACCESS_KEY_ID", "1280qwed9u9nsc9fdsbv9gsfrd")
				// os.Setenv("AWS_SECRET_ACCESS_KEY", "SED)SDVfdv0jfds08sdfgu09sd943tj4fELH/")
				// os.Setenv("AWS_SESSION_TOKEN", "IQoJb3JpZ2luX2VjELH//////////wEaCWV1LXdlc3QtMiJIMEYCIQDPU6UGJ0...df.fdgdfg.dfg.gdf.dgf")
				// os.Setenv("AWS_REGION", "eu-west-1")
				return func() {
					os.Clearenv()
				}
			},
		},
	}

	for name, tt := range ttests {
		t.Run(name, func(t *testing.T) {
			tearDown := tt.setupEnv()
			defer tearDown()
			impl, err := NewVaultStore(context.TODO(), tt.token, tt.conf)
			if err != nil {
				if err.Error() != tt.expect {
					t.Fatalf("failed to init hashivault, %v", err.Error())
				}
				return
			}

			impl.svc = tt.mockClient(t)
			rs := newRetrieveStrategy(NewDefatultStrategy(), tt.conf)
			rs.setImplementation(impl)
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

func TestAwsIamAuth(t *testing.T) {
	ttests := map[string]struct {
		token      string
		conf       GenVarsConfig
		expect     string
		mockClient func(t *testing.T) hashiVaultApi
		setupEnv   func() func()
	}{
		"test1": {
			// objType: nil,
		},
	}
	for name, tt := range ttests {
		t.Run(name, func(t *testing.T) {
			_ = tt
		})
	}
}
