package cmdutils_test

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/dnitsch/configmanager/internal/cmdutils"
	"github.com/dnitsch/configmanager/internal/config"
	"github.com/dnitsch/configmanager/internal/testutils"
	"github.com/dnitsch/configmanager/pkg/generator"
)

type mockCfgMgr struct {
	parsedMap    generator.ParsedMap
	err          error
	parsedString string
	config       *config.GenVarsConfig
}

func (m mockCfgMgr) RetrieveWithInputReplaced(input string) (string, error) {
	return m.parsedString, m.err
}

func (m mockCfgMgr) Retrieve(tokens []string) (generator.ParsedMap, error) {
	return m.parsedMap, m.err
}

func (m mockCfgMgr) GeneratorConfig() *config.GenVarsConfig {
	return m.config
}

func Test_UploadTokens_errors(t *testing.T) {
	m := &mockCfgMgr{}
	cmd := cmdutils.New(m)
	tokenMap := make(map[string]string)
	if err := cmd.UploadTokensWithVals(tokenMap); err == nil {
		t.Errorf(testutils.TestPhraseWithContext, "NOT YET IMPLEMENTED should fail", nil, "err")
	}
}

func cmdTestHelper(t *testing.T, err error, got []byte, expect []string) {
	t.Helper()
	if err != nil {
		t.Errorf("wanted file to not Error")
	}

	if len(got) < 1 {
		t.Error("empty file")
	}
	for _, want := range expect {
		if !strings.Contains(string(got), want) {
			t.Errorf(testutils.TestPhraseWithContext, "contents not found", string(got), want)
		}
	}
}

func Test_GenerateFromCmd(t *testing.T) {
	ttests := map[string]struct {
		mockMap generator.ParsedMap
		tokens  []string
		expect  []string
	}{
		"succeeds with 3 tokens": {
			generator.ParsedMap{"FOO://bar/qusx": "aksujg", "FOO://bar/lorem": "", "FOO://bar/ducks": "sdhbjk0293"},
			[]string{"FOO://bar/qusx", "FOO://bar/lorem", "FOO://bar/ducks"},
			[]string{"export QUSX='aksujg'", "export LOREM=''", "export DUCKS='sdhbjk0293'"},
		},
	}
	for name, tt := range ttests {
		t.Run(name, func(t *testing.T) {
			// create a temp file
			f, _ := os.CreateTemp(os.TempDir(), "gen-conf-frrom-token*")
			defer os.Remove(f.Name())

			m := &mockCfgMgr{
				config:    config.NewConfig(),
				parsedMap: tt.mockMap,
			}

			cmd := cmdutils.New(m)
			err := cmd.GenerateFromCmd(tt.tokens, f.Name())
			if err != nil {
				t.Fatalf(testutils.TestPhraseWithContext, "generate from cmd tokens", err, nil)
			}

			got, err := io.ReadAll(f)
			cmdTestHelper(t, err, got, tt.expect)
		})
	}
}

type mockWriter struct {
	w io.Writer
}

func (m *mockWriter) Close() error {
	return nil
}

func (m *mockWriter) Write(in []byte) (int, error) {
	return m.w.Write(in)
}

func Test_GenerateStrOut(t *testing.T) {

	inputStr := `FOO://bar/qusx FOO://bar/lorem FOO://bar/ducks`
	mockParsedStr := `aksujg fooLorem Mighty`
	expect := []string{"aksujg", "fooLorem", "Mighty"}

	t.Run("succeeds with input from string and output different", func(t *testing.T) {
		tearDown, reader, file := func(t *testing.T) (func(), io.Reader, string) {
			f, _ := os.CreateTemp(os.TempDir(), "gen-conf-frrom-string*")
			return func() {
				os.Remove(f.Name())
			}, f, f.Name()
		}(t)
		defer tearDown()
		m := &mockCfgMgr{
			config:       config.NewConfig(),
			parsedString: mockParsedStr,
		}
		cmd := cmdutils.New(m)
		err := cmd.GenerateStrOut(inputStr, file)
		if err != nil {
			t.Fatalf(testutils.TestPhraseWithContext, "generate from string", err, nil)
		}
		got, err := io.ReadAll(reader)
		cmdTestHelper(t, err, got, expect)
	})

	t.Run("succeeds output set to stdout", func(t *testing.T) {
		m := &mockCfgMgr{
			config:       config.NewConfig(),
			parsedString: mockParsedStr,
		}
		cmd := cmdutils.New(m)

		writer := bytes.NewBuffer([]byte{})
		mw := &mockWriter{w: writer}
		cmd.Writer = mw
		err := cmd.GenerateStrOut(inputStr, "stdout")
		if err != nil {
			t.Fatalf(testutils.TestPhraseWithContext, "generate from string", err, nil)
		}
		got, err := io.ReadAll(writer)
		cmdTestHelper(t, err, got, expect)
	})
	t.Run("succeeds input and output are set to file names", func(t *testing.T) {
		inputF, _ := os.CreateTemp(os.TempDir(), "gen-conf-frrom-string*")
		inputF.Write([]byte(inputStr))
		outputF, _ := os.CreateTemp(os.TempDir(), "gen-conf-frrom-string*")
		defer func() {
			os.Remove(inputF.Name())
			os.Remove(outputF.Name())
		}()

		m := &mockCfgMgr{
			config:       config.NewConfig(),
			parsedString: mockParsedStr,
		}
		cmd := cmdutils.New(m)
		err := cmd.GenerateStrOut(inputF.Name(), outputF.Name())
		if err != nil {
			t.Fatalf(testutils.TestPhraseWithContext, "generate from string", err, nil)
		}
		got, err := io.ReadAll(outputF)
		cmdTestHelper(t, err, got, expect)
	})

	t.Run("succeeds input and output are set to the same file", func(t *testing.T) {
		inputF, _ := os.CreateTemp(os.TempDir(), "gen-conf-frrom-string*")
		inputF.Write([]byte(inputStr))
		// outputF, _ := os.CreateTemp(os.TempDir(), "gen-conf-frrom-string*")
		defer func() {
			os.Remove(inputF.Name())
		}()

		m := &mockCfgMgr{
			config:       config.NewConfig().WithOutputPath(inputF.Name()),
			parsedString: mockParsedStr,
		}
		cmd := cmdutils.New(m)
		err := cmd.GenerateStrOut(inputF.Name(), inputF.Name())
		if err != nil {
			t.Fatalf(testutils.TestPhraseWithContext, "generate from string", err, nil)
		}
		got, err := os.ReadFile(inputF.Name())
		cmdTestHelper(t, err, got, expect)
	})
}

// import (
// 	"bytes"
// 	"fmt"
// 	"io"
// 	"os"
// 	"testing"

// 	"github.com/dnitsch/configmanager/internal/config"
// 	"github.com/dnitsch/configmanager/internal/testutils"
// 	"github.com/dnitsch/configmanager/pkg/generator"
// )

// type mockGenVars struct {
// 	mGen              func(tokens []string) (generator.ParsedMap, error)
// 	mConvertToExpVars func() []string
// 	mFlushToFile      func(w io.Writer, out []string) error
// 	config            *config.GenVarsConfig
// 	confOutputPath    string
// }

// var (
// 	tempOutPath = ""
// )

// func (m *mockGenVars) Generate(tokens []string) (generator.ParsedMap, error) {
// 	return m.mGen(tokens)
// }

// func (m *mockGenVars) ConvertToExportVar() []string {
// 	return m.mConvertToExpVars()
// }

// func (m *mockGenVars) FlushToFile(w io.Writer, out []string) error {
// 	return m.mFlushToFile(w, out)
// }

// func (m *mockGenVars) StrToFile(w io.Writer, str string) error {
// 	return generator.NewGenerator().StrToFile(w, str)
// }

// func (m *mockGenVars) Config() *config.GenVarsConfig {
// 	return m.config
// }

// func (m *mockGenVars) ConfigOutputPath() string {
// 	return m.confOutputPath
// }

// type mockRetrieveWithInput func(input string, config config.GenVarsConfig) (string, error)

// func (m mockRetrieveWithInput) RetrieveWithInputReplaced(input string, config config.GenVarsConfig) (string, error) {
// 	return m(input, config)
// }

// func Test_generateStrOutFromInput(t *testing.T) {
// 	tests := map[string]struct {
// 		confmgrMock func(t *testing.T) confMgrRetrieveWithInputReplacediface
// 		genMock     func(t *testing.T) generator.GenVarsiface
// 		w           func([]byte) io.Writer
// 		in          string
// 		expect      string
// 	}{
// 		"standard replace": {
// 			confmgrMock: func(t *testing.T) confMgrRetrieveWithInputReplacediface {
// 				return mockRetrieveWithInput(func(input string, config config.GenVarsConfig) (string, error) {
// 					return "pass=val1", nil
// 				})
// 			},
// 			genMock: func(t *testing.T) generator.GenVarsiface {
// 				gen := &mockGenVars{}
// 				gen.config = generator.NewConfig()

// 				return gen
// 			},
// 			in:     "pass=FOO#/test",
// 			expect: "pass=val1",
// 		},
// 	}
// 	for name, tt := range tests {
// 		t.Run(name, func(t *testing.T) {
// 			cu := &CmdUtils{
// 				cfgmgr:    tt.confmgrMock(t),
// 				generator: tt.genMock(t),
// 			}
// 			// want := []byte("pass=val1")
// 			in := bytes.NewBuffer([]byte(tt.in))
// 			out := bytes.NewBuffer([]byte{})
// 			if err := cu.generateStrOutFromInput(in, out); err != nil {
// 				t.Error(err)
// 			}
// 			got, err := io.ReadAll(out)
// 			if err != nil {
// 				t.Fatal(err)
// 			}
// 			if string(got) != tt.expect {
// 				t.Errorf(testutils.TestPhrase, string(got), tt.expect)
// 			}
// 		})
// 	}
// }

// func Test_generateFromStrOutOverwrite(t *testing.T) {
// 	tests := map[string]struct {
// 		confmgrMock func(t *testing.T) confMgrRetrieveWithInputReplacediface
// 		genMock     func(t *testing.T, out string) generator.GenVarsiface
// 		w           func([]byte) io.Writer
// 		in          string
// 		expect      string
// 	}{
// 		"standard replace": {
// 			confmgrMock: func(t *testing.T) confMgrRetrieveWithInputReplacediface {
// 				return mockRetrieveWithInput(func(input string, config config.GenVarsConfig) (string, error) {
// 					return "pass=val1", nil
// 				})
// 			},
// 			genMock: func(t *testing.T, out string) generator.GenVarsiface {
// 				gen := &mockGenVars{}
// 				gen.config = generator.NewConfig().WithOutputPath(out)
// 				return gen
// 			},
// 			in:     "pass=FOO#/test",
// 			expect: "pass=val1",
// 		},
// 	}
// 	for name, tt := range tests {
// 		t.Run(name, func(t *testing.T) {
// 			tempinfile, err := os.CreateTemp(os.TempDir(), "configmanager-mock-in")
// 			if err != nil {
// 				t.Fatal(err)
// 			}
// 			defer os.Remove(tempinfile.Name())
// 			if err := os.WriteFile(tempinfile.Name(), []byte(tt.in), 0644); err != nil {
// 				t.Fatal(err)
// 			}

// 			tempoutfile, err := os.CreateTemp(os.TempDir(), "configmanager-mock-out")
// 			if err != nil {
// 				t.Fatal(err)
// 			}
// 			defer os.Remove(tempoutfile.Name())

// 			cu := &CmdUtils{
// 				cfgmgr:    tt.confmgrMock(t),
// 				generator: tt.genMock(t, tempoutfile.Name()),
// 			}
// 			tempOutPath = tempoutfile.Name()
// 			outwriter, err := writer(tempoutfile.Name())
// 			if err != nil {
// 				t.Fatal(err)
// 			}
// 			if err := cu.generateFromStrOutOverwrite(tempinfile.Name(), tempoutfile.Name(), outwriter); err != nil {
// 				t.Fatal(err)
// 			}
// 			got, err := os.ReadFile(tempoutfile.Name())
// 			if err != nil {
// 				t.Fatal(err)
// 			}
// 			if string(got) != tt.expect {
// 				t.Errorf(testutils.TestPhrase, string(got), tt.expect)
// 			}
// 		})
// 	}
// }

// func TestGenerateStrOut(t *testing.T) {
// 	type testReturn struct {
// 		name      string
// 		isFile    bool
// 		inOutSame bool
// 	}
// 	ttests := map[string]struct {
// 		input       func() testReturn
// 		output      func() testReturn
// 		confmgrMock func(t *testing.T) confMgrRetrieveWithInputReplacediface
// 		genMock     func(t *testing.T, out string) generator.GenVarsiface
// 	}{
// 		"without overwrite": {
// 			input: func() testReturn {
// 				return testReturn{
// 					name:   "token",
// 					isFile: false,
// 				}
// 			},
// 			output: func() testReturn {
// 				return testReturn{
// 					name:   "replaced",
// 					isFile: false,
// 				}
// 			},
// 			confmgrMock: func(t *testing.T) confMgrRetrieveWithInputReplacediface {
// 				return mockRetrieveWithInput(func(input string, config config.GenVarsConfig) (string, error) {
// 					return "replaced", nil
// 				})
// 			},
// 			genMock: func(t *testing.T, out string) generator.GenVarsiface {
// 				gen := &mockGenVars{}
// 				gen.confOutputPath = out
// 				gen.config = generator.NewConfig().WithOutputPath(out)
// 				return gen
// 			},
// 		},
// 		"overwrite": {
// 			input: func() testReturn {
// 				tf, err := os.CreateTemp(os.TempDir(), "configmanager-mock-in")
// 				if err != nil {
// 					t.Fatal(err)
// 				}
// 				return testReturn{
// 					name:      tf.Name(),
// 					isFile:    true,
// 					inOutSame: true,
// 				}
// 			},
// 			output: func() testReturn {
// 				tf, err := os.CreateTemp(os.TempDir(), "configmanager-mock-out")
// 				if err != nil {
// 					t.Fatal(err)
// 				}
// 				return testReturn{
// 					name:   tf.Name(),
// 					isFile: true,
// 				}
// 			},
// 			confmgrMock: func(t *testing.T) confMgrRetrieveWithInputReplacediface {
// 				return mockRetrieveWithInput(func(input string, config config.GenVarsConfig) (string, error) {
// 					return "replaced", nil
// 				})
// 			},
// 			genMock: func(t *testing.T, out string) generator.GenVarsiface {
// 				gen := &mockGenVars{}
// 				gen.confOutputPath = out
// 				gen.config = generator.NewConfig().WithOutputPath(out)
// 				return gen
// 			},
// 		},
// 	}
// 	for name, tt := range ttests {
// 		t.Run(name, func(t *testing.T) {
// 			in, out := tt.input().name, tt.output().name
// 			defer os.Remove(in)
// 			defer os.Remove(tt.output().name)

// 			if tt.input().inOutSame {
// 				out = in
// 			}

// 			cu := &CmdUtils{
// 				cfgmgr:    tt.confmgrMock(t),
// 				generator: tt.genMock(t, out),
// 			}

// 			if err := cu.GenerateStrOut(in, out); err != nil {
// 				t.Errorf(testutils.TestPhrase, err, nil)
// 			}
// 			// if tt.input().isFile {
// 			// }
// 			// if tt.output().isFile {
// 			// }
// 		})
// 	}
// }

// func Test_generateFromToken(t *testing.T) {
// 	ttests := map[string]struct {
// 		confmgrMock func(t *testing.T) confMgrRetrieveWithInputReplacediface
// 		genMock     func(t *testing.T) generator.GenVarsiface
// 		tokens      []string
// 		w           func([]byte) io.Writer
// 		expect      string
// 	}{
// 		"success": {
// 			confmgrMock: func(t *testing.T) confMgrRetrieveWithInputReplacediface {
// 				return mockRetrieveWithInput(func(input string, config config.GenVarsConfig) (string, error) {
// 					return "replaced", nil
// 				})
// 			},
// 			genMock: func(t *testing.T) generator.GenVarsiface {
// 				gen := &mockGenVars{}
// 				gen.config = generator.NewConfig()
// 				gen.mGen = func(tokens []string) (generator.ParsedMap, error) {
// 					gm := generator.ParsedMap{}
// 					gm["foo"] = "replaced"
// 					if tokens[0] != "foo" {
// 						t.Errorf(testutils.TestPhrase, tokens[0], "foo")
// 					}
// 					return gm, nil
// 				}
// 				gen.mConvertToExpVars = func() []string {
// 					return []string{"export FOO=replaced"}
// 				}
// 				gen.mFlushToFile = func(w io.Writer, out []string) error {
// 					_, _ = w.Write([]byte("export FOO=replaced"))
// 					return nil
// 				}
// 				return gen
// 			},
// 			tokens: []string{"foo"},
// 			w: func(b []byte) io.Writer {
// 				return bytes.NewBuffer(b)
// 			},
// 			expect: "export FOO=replaced",
// 		},
// 		"error": {
// 			confmgrMock: func(t *testing.T) confMgrRetrieveWithInputReplacediface {
// 				return mockRetrieveWithInput(func(input string, config config.GenVarsConfig) (string, error) {
// 					return "replaced", nil
// 				})
// 			},
// 			genMock: func(t *testing.T) generator.GenVarsiface {
// 				gen := &mockGenVars{}
// 				gen.config = generator.NewConfig()
// 				gen.mGen = func(tokens []string) (generator.ParsedMap, error) {
// 					gm := generator.ParsedMap{}
// 					return gm, fmt.Errorf("unable to generate secrets from tokens")
// 				}
// 				return gen
// 			},
// 			tokens: []string{"foo"},
// 			w: func(b []byte) io.Writer {
// 				return bytes.NewBuffer(b)
// 			},
// 			expect: "unable to generate secrets from tokens",
// 		},
// 	}
// 	for name, tt := range ttests {
// 		t.Run(name, func(t *testing.T) {
// 			b := new(bytes.Buffer)
// 			cu := &CmdUtils{
// 				cfgmgr:    tt.confmgrMock(t),
// 				generator: tt.genMock(t),
// 			}

// 			if err := cu.generateFromToken(tt.tokens, b); err != nil {
// 				if err.Error() != tt.expect {
// 					t.Errorf(testutils.TestPhrase, err, nil)
// 				}
// 				return
// 			}
// 			out, _ := io.ReadAll(b)
// 			if len(out) < 1 {
// 				t.Errorf(testutils.TestPhrase, out, "to not be be empty")
// 			}
// 			if string(out) != tt.expect {
// 				t.Errorf(testutils.TestPhrase, string(out), tt.expect)
// 			}
// 		})
// 	}
// }

// func TestGenerateFromCmd(t *testing.T) {
// 	ttests := map[string]struct {
// 		confmgrMock func(t *testing.T) confMgrRetrieveWithInputReplacediface
// 		genMock     func(t *testing.T) generator.GenVarsiface
// 		tokens      []string
// 		output      func(t *testing.T) string
// 		expect      string
// 	}{
// 		"success to file": {
// 			func(t *testing.T) confMgrRetrieveWithInputReplacediface {
// 				return mockRetrieveWithInput(func(input string, config config.GenVarsConfig) (string, error) {
// 					return "replaced", nil
// 				})
// 			}, func(t *testing.T) generator.GenVarsiface {
// 				gen := &mockGenVars{}
// 				gen.config = generator.NewConfig()
// 				gen.mGen = func(tokens []string) (generator.ParsedMap, error) {
// 					gm := generator.ParsedMap{}
// 					gm["foo"] = "replaced"
// 					if tokens[0] != "foo" {
// 						t.Errorf(testutils.TestPhrase, tokens[0], "foo")
// 					}
// 					return gm, nil
// 				}
// 				gen.mConvertToExpVars = func() []string {
// 					return []string{"export FOO=replaced"}
// 				}
// 				gen.mFlushToFile = func(w io.Writer, out []string) error {
// 					_, _ = w.Write([]byte("export FOO=replaced"))
// 					return nil
// 				}
// 				return gen
// 			},
// 			[]string{"foo"},
// 			func(t *testing.T) string {
// 				tempoutfile, err := os.CreateTemp(os.TempDir(), "configmanager-mock-out")
// 				if err != nil {
// 					t.Fatal(err)
// 				}
// 				return tempoutfile.Name()
// 			},
// 			"export FOO=replaced",
// 		},
// 		// "success to stdout": {
// 		// 	func(t *testing.T) confMgrRetrieveWithInputReplacediface {
// 		// 		return mockRetrieveWithInput(func(input string, config config.GenVarsConfig) (string, error) {
// 		// 			return "replaced", nil
// 		// 		})
// 		// 	}, func(t *testing.T) generator.GenVarsiface {
// 		// 		gen := &mockGenVars{}
// 		// 		gen.config = generator.NewConfig()
// 		// 		gen.mGen = func(tokens []string) (generator.ParsedMap, error) {
// 		// 			gm := generator.ParsedMap{}
// 		// 			gm["foo"] = "replaced"
// 		// 			if tokens[0] != "foo" {
// 		// 				t.Errorf(testutils.TestPhrase, tokens[0], "foo")
// 		// 			}
// 		// 			return gm, nil
// 		// 		}
// 		// 		gen.mConvertToExpVars = func() []string {
// 		// 			return []string{"export FOO=replaced"}
// 		// 		}
// 		// 		gen.mFlushToFile = func(w io.Writer, out []string) error {
// 		// 			_, _ = w.Write([]byte("export FOO=replaced"))
// 		// 			return nil
// 		// 		}
// 		// 		return gen
// 		// 	},
// 		// 	[]string{"foo"},
// 		// 	func(t *testing.T) string {
// 		// 		return "stdout"
// 		// 	},
// 		// 	"export FOO=replaced",
// 		// },
// 	}
// 	for name, tt := range ttests {
// 		t.Run(name, func(t *testing.T) {
// 			outputPath := tt.output(t)
// 			if outputPath != "stdout" {
// 				defer os.Remove(outputPath)
// 			}

// 			cu := &CmdUtils{
// 				cfgmgr:    tt.confmgrMock(t),
// 				generator: tt.genMock(t),
// 			}
// 			if err := cu.GenerateFromCmd(tt.tokens, outputPath); err != nil {
// 				if err.Error() != tt.expect {
// 					t.Errorf(testutils.TestPhrase, err.Error(), tt.expect)
// 				}
// 				return
// 			}
// 		})
// 	}
// }

// // func Test_generateFromStrOut(t *testing.T) {
// // 	ttests := map[string]struct {
// // 		objType	any

// // 	}{
// // 		"test1":
// // 		{
// // 			objType: nil,

// // 		},
// // 	}
// // 	for name, tt := range ttests {
// // 		t.Run(name, func(t *testing.T) {

// // 		})
// // 	}
// // }
