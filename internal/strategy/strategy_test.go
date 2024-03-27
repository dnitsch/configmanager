package strategy_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/dnitsch/configmanager/internal/config"
	"github.com/dnitsch/configmanager/internal/store"
	"github.com/dnitsch/configmanager/internal/strategy"
	"github.com/dnitsch/configmanager/internal/testutils"
	"github.com/go-test/deep"
)

type mockGenerate struct {
	inToken, value string
	err            error
}

func (m mockGenerate) SetToken(s *config.ParsedTokenConfig) {
}

func (m mockGenerate) Token() (s string, e error) {
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
			"AZTABLESTORE://mountPath/token",
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
			rs := strategy.New(store.NewDefatultStrategy(), *tt.config)
			token, _ := config.NewParsedTokenConfig(tt.token, *tt.config)
			got := rs.RetrieveByToken(context.TODO(), tt.impl(t), token)
			if got.Err != nil {
				t.Errorf(testutils.TestPhraseWithContext, "Token response errored", got.Err.Error(), tt.expect)
			}
			if got.Value() != tt.expect {
				t.Errorf(testutils.TestPhraseWithContext, "Value not correct", got.Value(), tt.expect)
			}
			if got.Key().String() != tt.token {
				t.Errorf(testutils.TestPhraseWithContext, "INcorrect Token returned in Key", got.Key().String(), tt.token)
			}
		})
	}
}

func Test_CustomStrategyFuncMap_add_own(t *testing.T) {
	// t.Skip()
	ttests := map[string]struct {
	}{
		"default": {},
	}
	for name, _ := range ttests {
		t.Run(name, func(t *testing.T) {
			called := 0
			genVarsConf := config.NewConfig()
			token, _ := config.NewParsedTokenConfig("AZTABLESTORE://mountPath/token", *genVarsConf)

			var custFunc = func(ctx context.Context, token *config.ParsedTokenConfig) (store.Strategy, error) {
				m := &mockGenerate{"AZTABLESTORE://mountPath/token", "bar", nil}
				called++
				return m, nil
			}

			s := strategy.New(store.NewDefatultStrategy(), *genVarsConf)
			s.WithStrategyFuncMap(strategy.StrategyFuncMap{config.AzTableStorePrefix: custFunc})

			store, _ := s.SelectImplementation(context.TODO(), token)
			_ = s.RetrieveByToken(context.TODO(), store, token)

			if called != 1 {
				t.Errorf(testutils.TestPhraseWithContext, "custom func not called", called, 1)
			}
		})
	}
}

func Test_SelectImpl_With(t *testing.T) {

	ttests := map[string]struct {
		setUpTearDown func() func()
		token         string
		config        *config.GenVarsConfig
		expect        func() store.Strategy
		expErr        error
	}{
		"unknown": {
			func() func() {
				return func() {
				}
			},
			"UNKNOWN#foo/bar",
			config.NewConfig().WithTokenSeparator("#"),
			func() store.Strategy { return nil },
			fmt.Errorf("implementation not found for input string: UNKNOWN#foo/bar"),
		},
		"success AZTABLESTORE": {
			func() func() {
				os.Setenv("AZURE_stuff", "foo")
				return func() {
					os.Clearenv()
				}
			},
			"AZTABLESTORE#foo/bar1",
			config.NewConfig().WithTokenSeparator("#"),
			func() store.Strategy {
				conf, _ := config.NewParsedTokenConfig("AZTABLESTORE#foo/bar1", *config.NewConfig().WithTokenSeparator("#"))
				s, _ := store.NewAzTableStore(context.TODO(), conf)
				return s
			},
			nil,
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
			want := tt.expect()
			rs := strategy.New(store.NewDefatultStrategy(), *tt.config)
			token, _ := config.NewParsedTokenConfig(tt.token, *tt.config)
			got, err := rs.SelectImplementation(context.TODO(), token)

			if err != nil {
				if err.Error() != tt.expErr.Error() {
					t.Errorf(testutils.TestPhraseWithContext, "uncaught error", err.Error(), tt.expErr.Error())
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
