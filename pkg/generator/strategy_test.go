package generator_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/dnitsch/configmanager/internal/config"
	"github.com/dnitsch/configmanager/internal/store"
	"github.com/dnitsch/configmanager/internal/testutils"
	"github.com/dnitsch/configmanager/pkg/generator"
	"github.com/go-test/deep"
)

type mockGenerate struct {
	inToken, value string
	err            error
}

func (m *mockGenerate) SetToken(s *config.ParsedTokenConfig) {
}
func (m *mockGenerate) Token() (s string, e error) {
	return m.value, m.err
}

func Test_Strategy_Retrieve_succeeds(t *testing.T) {
	ttests := map[string]struct {
		impl   func(t *testing.T) store.Strategy
		config *config.GenVarsConfig
		token  string
		expect string
	}{
		"with mocked implementation AZTABLESTORAGE": {
			func(t *testing.T) store.Strategy {
				return &mockGenerate{"AZTABLESTORE://mountPath/token", "bar", nil}
			},
			config.NewConfig().WithOutputPath("stdout").WithTokenSeparator("://"),
			"AZTABLESTORE://token",
			"bar",
		},
		// "error in retrieval": {
		// 	func(t *testing.T) store.Strategy {
		// 		return &mockGenerate{"SOME://mountPath/token", "bar", fmt.Errorf("unable to perform getTokenValue")}
		// 	},
		// 	config.NewConfig().WithOutputPath("stdout").WithTokenSeparator("://"),
		// 	[]string{"SOME://token"},
		// 	config.AzAppConfigPrefix,
		// 	"unable to perform getTokenValue",
		// },
	}
	for name, tt := range ttests {
		t.Run(name, func(t *testing.T) {
			// channelResp := make(chan *generator.ChanResp)
			rs := generator.NewRetrieveStrategy(store.NewDefatultStrategy(), *tt.config)

			// var wg sync.WaitGroup
			got := rs.RetrieveByToken(context.TODO(), tt.impl(t), config.NewParsedTokenConfig(tt.token, *tt.config))
			if got.Err != nil {
				t.Errorf(testutils.TestPhraseWithContext, "Token response errored", got.Err.Error(), tt.expect)
			}
			// wg.Add(len(tt.token))
			// go func() {
			// 	defer wg.Done()
			// 	channelResp <- rs.RetrieveByToken(context.TODO(), tt.impl(t), config.NewParsedTokenConfig(tt.token[0], *tt.config))
			// }()
			// // var got *generator.ChanResp
			// // for {
			// // 	// channelResp := make(chan *generator.ChanResp)
			// // 	<-channelResp
			// // 	// select {
			// // 	// case resp := <-channelResp:
			// // 	// 	got = resp
			// // 	// 	return
			// // 	// }

			// // }

			// // got.
			// go func() {
			// 	wg.Wait()
			// 	close(channelResp)
			// }()
			// for g := range channelResp {
			// 	if g.err != nil {
			// 		if g.err.Error() != tt.expect {
			// 			t.Errorf(testutils.TestPhraseWithContext, "channel errored not expected", g.err.Error(), tt.expect)
			// 		}
			// 		return
			// 	}
			// 	if g.value != tt.expect {
			// 		t.Errorf(testutils.TestPhraseWithContext, "channel value", g.value, tt.expect)
			// 	}
			// }
		})
	}
}

var UnknownPrefix config.ImplementationPrefix = "WRONG"

func Test_SelectImpl_(t *testing.T) {
	ttests := map[string]struct {
		setUpTearDown func() func()
		token         string
		config        *config.GenVarsConfig
		expect        store.Strategy
	}{

		"success AZTABLESTORE": {
			func() func() {
				os.Setenv("AZURE_stuff", "foo")
				return func() {
					os.Clearenv()
				}
			},
			"AZTABLESTORE#foo/bar",
			config.NewConfig(),
			&store.AzTableStore{},
		},
		// "default Error": {
		// 	func() func() {
		// 		os.Setenv("AWS_ACCESS_KEY", "AAAAAAAAAAAAAAA")
		// 		os.Setenv("AWS_SECRET_ACCESS_KEY", "00000000000000000000111111111")
		// 		return func() {
		// 			os.Clearenv()
		// 		}
		// 	},
		// 	context.TODO(),
		// 	UnknownPrefix, "AWSPARAMSTR://foo/bar", (&GenVarsConfig{}).WithKeySeparator("|").WithTokenSeparator("://"),
		// 	func(t *testing.T, ctx context.Context, conf GenVarsConfig) genVarsStrategy {
		// 		imp, err := NewParamStore(ctx)
		// 		if err != nil {
		// 			t.Errorf(testutils.TestPhraseWithContext, "init impl error", err.Error(), nil)
		// 		}
		// 		return imp
		// 	},
		// },
	}
	for name, tt := range ttests {
		t.Run(name, func(t *testing.T) {
			tearDown := tt.setUpTearDown()
			defer tearDown()
			rs := generator.NewRetrieveStrategy(store.NewDefatultStrategy(), *tt.config)
			got, err := rs.SelectImplementation(context.TODO(), config.NewParsedTokenConfig(tt.token, *config.NewConfig().WithTokenSeparator("#")))

			if err != nil {
				if err.Error() != fmt.Sprintf("implementation not found for input string: %s", tt.token) {
					t.Errorf(testutils.TestPhraseWithContext, "uncaught error", err.Error(), fmt.Sprintf("implementation not found for input string: %s", tt.token))
				}
				return
			}

			diff := deep.Equal(got, tt.expect)
			if diff != nil {
				t.Errorf(testutils.TestPhraseWithContext, "reflection of initialised implentations", fmt.Sprintf("%q", got), fmt.Sprintf("%q", tt.expect))
			}
		})
	}
}

// "success AWSSEcretsMgr": {
// 	func() func() {
// 		os.Setenv("AWS_ACCESS_KEY", "AAAAAAAAAAAAAAA")
// 		os.Setenv("AWS_SECRET_ACCESS_KEY", "00000000000000000000111111111")
// 		return func() {
// 			os.Clearenv()
// 		}
// 	},
// 	context.TODO(),
// 	SecretMgrPrefix, "AWSSECRETS://foo/bar", (&GenVarsConfig{}).WithKeySeparator("|").WithTokenSeparator("://"),
// 	func(t *testing.T, ctx context.Context, conf GenVarsConfig) store.Strategy {
// 		imp, err := NewSecretsMgr(ctx)
// 		if err != nil {
// 			t.Errorf(testutils.TestPhraseWithContext, "aws secrets init impl error", err.Error(), nil)
// 		}
// 		return imp
// 	},
// },
// "success AWSParamStore": {
// 	func() func() {
// 		os.Setenv("AWS_ACCESS_KEY", "AAAAAAAAAAAAAAA")
// 		os.Setenv("AWS_SECRET_ACCESS_KEY", "00000000000000000000111111111")
// 		return func() {
// 			os.Clearenv()
// 		}
// 	},
// 	context.TODO(),
// 	ParamStorePrefix, "AWSPARAMSTR://foo/bar", (&GenVarsConfig{}).WithKeySeparator("|").WithTokenSeparator("://"),
// 	func(t *testing.T, ctx context.Context, conf GenVarsConfig) genVarsStrategy {
// 		imp, err := NewParamStore(ctx)
// 		if err != nil {
// 			t.Errorf(testutils.TestPhraseWithContext, "paramstore init impl error", err.Error(), nil)
// 		}
// 		return imp
// 	},
// },
// "success GCPSecrets": {
// 	func() func() {
// 		tmp, _ := os.CreateTemp(".", "gcp-creds-*")
// 		tmp.Write(TEST_GCP_CREDS)
// 		os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", tmp.Name())
// 		return func() {
// 			os.Clearenv()
// 			os.Remove(tmp.Name())
// 		}
// 	},
// 	context.TODO(),
// 	GcpSecretsPrefix, "GCPSECRETS://foo/bar", (&GenVarsConfig{}).WithKeySeparator("|").WithTokenSeparator("://"),
// 	func(t *testing.T, ctx context.Context, conf GenVarsConfig) genVarsStrategy {
// 		imp, err := NewGcpSecrets(ctx)
// 		if err != nil {
// 			t.Errorf(testutils.TestPhraseWithContext, "gcp secrets init impl error", err.Error(), nil)
// 		}
// 		return imp
// 	},
// },
// "success AZKV": {
// 	func() func() {
// 		os.Setenv("AZURE_STUFF", "foo")
// 		return func() {
// 			os.Clearenv()
// 		}
// 	},
// 	context.TODO(),
// 	AzKeyVaultSecretsPrefix, "AZKVSECRET://foo/bar", (&GenVarsConfig{}).WithKeySeparator("|").WithTokenSeparator("://"),
// 	func(t *testing.T, ctx context.Context, conf GenVarsConfig) genVarsStrategy {
// 		imp, err := NewKvScrtStore(ctx, "AZKVSECRET://foo/bar", conf)
// 		if err != nil {
// 			t.Errorf(testutils.TestPhraseWithContext, "azkv init impl error", err.Error(), nil)
// 		}
// 		return imp
// 	},
// },
// "success Vault": {
// 	func() func() {
// 		os.Setenv("VAULT_TOKEN", "foo")
// 		os.Setenv("VAULT_ADDR", "http://127.0.0.1:8200")
// 		return func() {
// 			os.Clearenv()
// 		}
// 	},
// 	context.TODO(),
// 	HashicorpVaultPrefix, "VAULT://foo/bar", (&GenVarsConfig{}).WithKeySeparator("|").WithTokenSeparator("://"),
// 	func(t *testing.T, ctx context.Context, conf GenVarsConfig) genVarsStrategy {
// 		imp, err := NewVaultStore(ctx, "VAULT://foo/bar", conf)
// 		if err != nil {
// 			t.Errorf(testutils.TestPhraseWithContext, "vault init impl error", err.Error(), nil)
// 		}
// 		return imp
// 	},
// },
