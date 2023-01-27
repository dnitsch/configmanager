package generator

import (
	"context"
	"fmt"
	"os"
	"sync"
	"testing"

	"github.com/dnitsch/configmanager/internal/testutils"
	"github.com/go-test/deep"
)

type mockGenerate struct {
	token, value string
	err          error
}

func (m *mockGenerate) setToken(s string) {
}
func (m *mockGenerate) getTokenValue(rs *retrieveStrategy) (s string, e error) {
	return m.value, m.err
}

func Test_rsRetrieve(t *testing.T) {
	ttests := map[string]struct {
		impl       func(t *testing.T) genVarsStrategy
		config     GenVarsConfig
		token      []string
		implPrefix ImplementationPrefix
		expect     string
	}{
		"success retrieval": {
			func(t *testing.T) genVarsStrategy {
				return &mockGenerate{token: "SOME://mountPath/token", value: "bar", err: nil}
			},
			GenVarsConfig{keySeparator: "|", tokenSeparator: "://", outpath: "stdout"},
			[]string{"SOME://token"},
			HashicorpVaultPrefix,
			"bar",
		}, "error in retrieval": {
			func(t *testing.T) genVarsStrategy {
				return &mockGenerate{token: "SOME://mountPath/token", value: "bar", err: fmt.Errorf("unable to perform getTokenValue")}
			},
			GenVarsConfig{keySeparator: "|", tokenSeparator: "://", outpath: "stdout"},
			[]string{"SOME://token"},
			HashicorpVaultPrefix,
			"unable to perform getTokenValue",
		},
	}
	for name, tt := range ttests {
		t.Run(name, func(t *testing.T) {
			got := make(chan chanResp, len(tt.token))
			rs := &retrieveStrategy{NewDefatultStrategy(), tt.config, ""}
			var wg sync.WaitGroup

			wg.Add(len(tt.token))
			go func() {
				defer wg.Done()
				got <- rs.RetrieveByToken(context.TODO(), tt.impl(t), tt.implPrefix, tt.token[0])
			}()

			go func() {
				wg.Wait()
				close(got)
			}()
			for g := range got {
				if g.err != nil {
					if g.err.Error() != tt.expect {
						t.Errorf(testutils.TestPhraseWithContext, "channel errored not expected", g.err.Error(), tt.expect)
					}
					return
				}
				if g.value != tt.expect {
					t.Errorf(testutils.TestPhraseWithContext, "channel value", g.value, tt.expect)
				}
			}
		})
	}
}

var UnknownPrefix ImplementationPrefix = "WRONG"

func TestSelectImpl(t *testing.T) {
	ttests := map[string]struct {
		setUpTearDown func() func()
		ctx           context.Context
		prefix        ImplementationPrefix
		in            string
		config        *GenVarsConfig
		expect        func(t *testing.T, ctx context.Context) genVarsStrategy
	}{
		"success AWSSEcretsMgr": {
			func() func() {
				os.Setenv("AWS_PROFILE", "foo")
				return func() {
					os.Clearenv()
				}
			},
			context.TODO(),
			SecretMgrPrefix, "AWSSECRETS://foo/bar", (&GenVarsConfig{}).WithKeySeparator("|").WithTokenSeparator("://"),
			func(t *testing.T, ctx context.Context) genVarsStrategy {
				imp, err := NewSecretsMgr(ctx)
				if err != nil {
					t.Errorf(testutils.TestPhraseWithContext, "aws secrets init impl error", err.Error(), nil)
				}
				return imp
			},
		},
		"success AWSParamStore": {
			func() func() {
				os.Setenv("AWS_PROFILE", "foo")
				return func() {
					os.Clearenv()
				}
			},
			context.TODO(),
			ParamStorePrefix, "AWSPARAMSTR://foo/bar", (&GenVarsConfig{}).WithKeySeparator("|").WithTokenSeparator("://"),
			func(t *testing.T, ctx context.Context) genVarsStrategy {
				imp, err := NewParamStore(ctx)
				if err != nil {
					t.Errorf(testutils.TestPhraseWithContext, "paramstore init impl error", err.Error(), nil)
				}
				return imp
			},
		},
		"success GCPSecrets": {
			func() func() {
				tmp, _ := os.CreateTemp(".", "gcp-creds-*")
				tmp.Write(TEST_GCP_CREDS)
				os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", tmp.Name())
				return func() {
					os.Clearenv()
					os.Remove(tmp.Name())
				}
			},
			context.TODO(),
			GcpSecretsPrefix, "GCPSECRETS://foo/bar", (&GenVarsConfig{}).WithKeySeparator("|").WithTokenSeparator("://"),
			func(t *testing.T, ctx context.Context) genVarsStrategy {
				imp, err := NewGcpSecrets(ctx)
				if err != nil {
					t.Errorf(testutils.TestPhraseWithContext, "gcp secrets init impl error", err.Error(), nil)
				}
				return imp
			},
		},
		"success AZKV": {
			func() func() {
				os.Setenv("AZURE_STUFF", "foo")
				return func() {
					os.Clearenv()
				}
			},
			context.TODO(),
			AzKeyVaultSecretsPrefix, "AZKVSECRET://foo/bar", (&GenVarsConfig{}).WithKeySeparator("|").WithTokenSeparator("://"),
			func(t *testing.T, ctx context.Context) genVarsStrategy {
				imp, err := NewKvScrtStore(ctx, "AZKVSECRET://foo/bar", "://", "|")
				if err != nil {
					t.Errorf(testutils.TestPhraseWithContext, "azkv init impl error", err.Error(), nil)
				}
				return imp
			},
		},
		"success Vault": {
			func() func() {
				os.Setenv("VAULT_TOKEN", "foo")
				os.Setenv("VAULT_ADDR", "http://127.0.0.1:8200")
				return func() {
					os.Clearenv()
				}
			},
			context.TODO(),
			HashicorpVaultPrefix, "VAULT://foo/bar", (&GenVarsConfig{}).WithKeySeparator("|").WithTokenSeparator("://"),
			func(t *testing.T, ctx context.Context) genVarsStrategy {
				imp, err := NewVaultStore(ctx, "VAULT://foo/bar", "://", "|")
				if err != nil {
					t.Errorf(testutils.TestPhraseWithContext, "vault init impl error", err.Error(), nil)
				}
				return imp
			},
		},
		"default Error": {
			func() func() {
				os.Setenv("AWS_PROFILE", "foo")
				return func() {
					os.Clearenv()
				}
			},
			context.TODO(),
			UnknownPrefix, "AWSPARAMSTR://foo/bar", (&GenVarsConfig{}).WithKeySeparator("|").WithTokenSeparator("://"),
			func(t *testing.T, ctx context.Context) genVarsStrategy {
				imp, err := NewParamStore(ctx)
				if err != nil {
					t.Errorf(testutils.TestPhraseWithContext, "init impl error", err.Error(), nil)
				}
				return imp
			},
		},
	}
	for name, tt := range ttests {
		t.Run(name, func(t *testing.T) {
			tearDown := tt.setUpTearDown()
			defer tearDown()
			rs := &retrieveStrategy{}
			want := tt.expect(t, tt.ctx)
			got, err := rs.SelectImplementation(tt.ctx, tt.prefix, tt.in, tt.config)
			if err != nil {
				if err.Error() != fmt.Sprintf("implementation not found for input string: %s", tt.in) {
					t.Errorf(testutils.TestPhraseWithContext, "uncaught error", err.Error(), fmt.Sprintf("implementation not found for input string: %s", tt.in))
				}
				return
			}

			diff := deep.Equal(got, want)
			if diff != nil {
				t.Errorf(testutils.TestPhraseWithContext, "reflection of initialised implentations", fmt.Sprintf("%q", got), fmt.Sprintf("%q", want))
			}
		})
	}
}
