package cmdutils

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"testing"

	"github.com/dnitsch/configmanager/internal/testutils"
	"github.com/dnitsch/configmanager/pkg/generator"
)

type mockGenVars struct{}

var (
	testKey = "FOO#/test"
	testVal = "val1"
)

func (m *mockGenVars) Generate(tokens []string) (generator.ParsedMap, error) {
	pm := generator.ParsedMap{}
	pm[testKey] = testVal
	return pm, nil
}

func (m *mockGenVars) ConvertToExportVar() []string {
	return []string{}
}

func (m *mockGenVars) FlushToFile(w io.Writer) error {
	return generator.NewGenerator().StrToFile(w, "pass='val1'")
}

func (m *mockGenVars) StrToFile(w io.Writer, str string) error {
	return generator.NewGenerator().StrToFile(w, str)
}

func (m *mockGenVars) Config() *generator.GenVarsConfig {
	return &generator.GenVarsConfig{}
}

func (m *mockGenVars) ConfigOutputPath() string {
}

type mockConfMgr struct{}

func (m *mockConfMgr) Retrieve(tokens []string, config generator.GenVarsConfig) (generator.ParsedMap, error) {
	pm := generator.ParsedMap{}
	pm[testKey] = testVal
	return pm, nil
}

func (m *mockConfMgr) RetrieveWithInputReplaced(input string, config generator.GenVarsConfig) (string, error) {
	return `pass=val1`, nil
}

func (m *mockConfMgr) Insert(force bool) error {
	return fmt.Errorf("unimplemented")
}

func Test_generateStrOutFromInput(t *testing.T) {
	tests := []struct {
		name     string
		cmdUtils *CmdUtils
	}{
		{
			name: "standard replace",
			cmdUtils: &CmdUtils{
				cfgmgr:    &mockConfMgr{},
				generator: &mockGenVars{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			want := []byte("pass=val1")
			in := bytes.NewBuffer([]byte("pass=FOO#/test"))
			out := bytes.NewBuffer([]byte{})
			if err := tt.cmdUtils.generateStrOutFromInput(in, out); err != nil {
				t.Error(err)
			}
			got, err := ioutil.ReadAll(out)
			if err != nil {
				t.Fatal(err)
			}
			if string(got) != string(want) {
				t.Errorf(testutils.TestPhrase, string(want), string(got))
			}
		})
	}
}

func Test_generateFromStrOutOverwrite(t *testing.T) {
	tests := []struct {
		name     string
		cmdUtils *CmdUtils
	}{
		{
			name: "standard replace",
			cmdUtils: &CmdUtils{
				cfgmgr:    &mockConfMgr{},
				generator: &mockGenVars{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			want := []byte("pass=val1")
			tempinfile, err := ioutil.TempFile(os.TempDir(), "configmanager-mock-in")
			if err != nil {
				t.Fatal(err)
			}
			defer os.Remove(tempinfile.Name())
			if err := os.WriteFile(tempinfile.Name(), []byte("pass=FOO#/test"), 0644); err != nil {
				t.Fatal(err)
			}

			tempoutfile, err := ioutil.TempFile(os.TempDir(), "configmanager-mock-out")
			if err != nil {
				t.Fatal(err)
			}
			defer os.Remove(tempoutfile.Name())

			outwriter, err := os.Open(tempoutfile.Name())
			if err != nil {
				t.Fatal(err)
			}
			if err := tt.cmdUtils.generateFromStrOutOverwrite(tempinfile.Name(), tempoutfile.Name(), outwriter); err != nil {
				t.Fatal(err)
			}
			got, err := ioutil.ReadAll(outwriter)
			if err != nil {
				t.Fatal(err)
			}
			if string(got) != string(want) {
				t.Errorf(testutils.TestPhrase, string(want), string(got))
			}
		})
	}
}
