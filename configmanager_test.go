package configmanager

import (
	"fmt"
	"io"
	"reflect"
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
	return nil
}

func (m *mockGenVars) StrToFile(w io.Writer, str string) error {
	return nil
}

func Test_retrieve(t *testing.T) {
	tests := []struct {
		name      string
		tokens    []string
		genvar    generator.Generatoriface
		expectKey string
		expectVal string
	}{
		{
			name:      "standard",
			tokens:    []string{"FOO#/test"},
			genvar:    &mockGenVars{},
			expectKey: testKey,
			expectVal: testVal,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pm, err := retrieve(tt.tokens, tt.genvar)
			if err != nil {
				t.Errorf(testutils.TestPhrase, nil, err)
			}
			for k, v := range pm {
				if k != tt.expectKey {
					t.Errorf(testutils.TestPhrase, tt.expectKey, k)
				}
				if v != tt.expectVal {
					t.Errorf(testutils.TestPhrase, tt.expectVal, k)
				}
			}
		})
	}
}

var (
	strT1 = `
space: preserved
	indents: preserved
	arr: [ "FOO#/test" ]
	// comments preserved
	arr:
		- "FOO#/test"
`
	strT2 = `
// TOML
[[somestuff]]
key = "FOO#/test" 
`

	strT3 = `
// TOML
[[somestuff]]
key = FOO#/test
key2 = FOO#/test
key3 = FOO#/test
key4 = FOO#/test
`

	strT4 = `
export FOO='FOO#/test'
export FOO1=FOO#/test
export FOO2='FOO#/test'
export FOO3=FOO#/test
export FOO4=FOO#/test

[[section]]

foo23 = FOO#/test
`
)

func Test_retrieveWithInputReplaced(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		genvar generator.Generatoriface
		expect string
	}{
		{
			name:   "strYaml",
			input:  strT1,
			genvar: &mockGenVars{},
			expect: `
space: preserved
	indents: preserved
	arr: [ "val1" ]
	// comments preserved
	arr:
		- "val1"
`,
		},
		{
			name:   "strToml",
			input:  strT2,
			genvar: &mockGenVars{},
			expect: `
// TOML
[[somestuff]]
key = "val1" 
`,
		},
		{
			name:   "strTomlWithoutQuotes",
			input:  strT3,
			genvar: &mockGenVars{},
			expect: `
// TOML
[[somestuff]]
key = val1
key2 = val1
key3 = val1
key4 = val1
`,
		},
		{
			name:   "strTomlWithoutMultiline",
			input:  strT4,
			genvar: &mockGenVars{},
			expect: `
export FOO='val1'
export FOO1=val1
export FOO2='val1'
export FOO3=val1
export FOO4=val1

[[section]]

foo23 = val1
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := retrieveWithInputReplaced(tt.input, tt.genvar)
			if err != nil {
				t.Errorf("failed with %v", err)
			}
			if got != tt.expect {
				t.Errorf(testutils.TestPhrase, tt.expect, got)
			}
		})
	}
}

func Test_replaceString(t *testing.T) {
	tests := []struct {
		name      string
		parsedMap generator.ParsedMap
		inputStr  string
		expectStr string
	}{
		{
			name: "ordered correctly",
			parsedMap: generator.ParsedMap{
				"AZKVSECRET#/test-vault/db-config|user": "foo",
				"AZKVSECRET#/test-vault/db-config|pass": "bar",
				"AZKVSECRET#/test-vault/db-config":      fmt.Sprintf("%v", "{\"user\": \"foo\", \"pass\": \"bar\"}"),
			},
			inputStr: `app: foo
db2: AZKVSECRET#/test-vault/db-config
db: 
	user: AZKVSECRET#/test-vault/db-config|user
	pass: AZKVSECRET#/test-vault/db-config|pass
`,
			expectStr: `app: foo
db2: {"user": "foo", "pass": "bar"}
db: 
	user: foo
	pass: bar
`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := replaceString(tt.parsedMap, tt.inputStr)
			if got != tt.expectStr {
				t.Errorf(testutils.TestPhrase, tt.expectStr, got)
			}
		})
	}
}

type MockCfgMgr struct {
	RetrieveWithInputReplacedTest func(input string, config generator.GenVarsConfig) (string, error)
}

func (m *MockCfgMgr) RetrieveWithInputReplaced(input string, config generator.GenVarsConfig) (string, error) {
	if m.RetrieveWithInputReplacedTest != nil {
		return m.RetrieveWithInputReplacedTest(input, config)
	}
	return "", nil
}

func (m *MockCfgMgr) Insert(force bool) error {
	return nil
}

func (m *MockCfgMgr) Retrieve(tokens []string, config generator.GenVarsConfig) (generator.ParsedMap, error) {
	return nil, nil
}

type testSimpleStruct struct {
	Foo string `json:"foo"`
	Bar string `json:"bar"`
}

type testAnotherNEst struct {
	Number int     `json:"number,omitempty"`
	Float  float32 `json:"float,omitempty"`
}

type testLol struct {
	Bla     string          `json:"bla,omitempty"`
	Another testAnotherNEst `json:"another,omitempty"`
}

type testNestedStruct struct {
	Foo string  `json:"foo"`
	Bar string  `json:"bar"`
	Lol testLol `json:"lol,omitempty"`
}

func Test_KubeControllerSpecHelper(t *testing.T) {
	tests := []struct {
		name     string
		testType testSimpleStruct
		expect   testSimpleStruct
		cfmgr    func(t *testing.T) ConfigManageriface
	}{
		{
			name: "happy path simple struct",
			testType: testSimpleStruct{
				Foo: "AWSSECRETS:///bar/foo",
				Bar: "quz",
			},
			expect: testSimpleStruct{
				Foo: "baz",
				Bar: "quz",
			},
			cfmgr: func(t *testing.T) ConfigManageriface {
				mcm := &MockCfgMgr{}
				mcm.RetrieveWithInputReplacedTest = func(input string, config generator.GenVarsConfig) (string, error) {
					return `{"foo":"baz","bar":"quz"}`, nil
				}
				return mcm
			},
		},
		{
			name: "happy path simple struct2",
			testType: testSimpleStruct{
				Foo: "AWSSECRETS:///bar/foo2",
				Bar: "quz",
			},
			expect: testSimpleStruct{
				Foo: "baz2",
				Bar: "quz",
			},
			cfmgr: func(t *testing.T) ConfigManageriface {
				mcm := &MockCfgMgr{}
				mcm.RetrieveWithInputReplacedTest = func(input string, config generator.GenVarsConfig) (string, error) {
					return `{"foo":"baz2","bar":"quz"}`, nil
				}
				return mcm
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			config := generator.NewConfig()
			resp, err := KubeControllerSpecHelper(tt.testType, tt.cfmgr(t), *config)
			if err != nil {
				t.Errorf("expected error to be <nil>, got: %v", err)
			}
			if !reflect.DeepEqual(resp, &tt.expect) {
				t.Error("")
			}
		})
	}
}

func Test_KubeControllerComplex(t *testing.T) {
	tests := []struct {
		name     string
		testType testNestedStruct
		expect   testNestedStruct
		cfmgr    func(t *testing.T) ConfigManageriface
	}{
		{
			name: "happy path simple struct",
			testType: testNestedStruct{
				Foo: "AWSSECRETS:///bar/foo",
				Bar: "quz",
				Lol: testLol{
					Bla: "booo",
					Another: testAnotherNEst{
						Number: 1235,
						Float:  123.09,
					},
				},
			},
			expect: testNestedStruct{
				Foo: "baz",
				Bar: "quz",
				Lol: testLol{
					Bla: "booo",
					Another: testAnotherNEst{
						Number: 1235,
						Float:  123.09,
					},
				},
			},
			cfmgr: func(t *testing.T) ConfigManageriface {
				mcm := &MockCfgMgr{}
				mcm.RetrieveWithInputReplacedTest = func(input string, config generator.GenVarsConfig) (string, error) {
					return `{"foo":"baz","bar":"quz", "lol":{"bla":"booo","another":{"number": 1235, "float": 123.09}}}`, nil
				}
				return mcm
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := generator.NewConfig()
			resp, err := KubeControllerSpecHelper(tt.testType, tt.cfmgr(t), *config)
			if err != nil {
				t.Errorf("expected error to be <nil>, got: %v", err)
			}
			if !reflect.DeepEqual(resp, &tt.expect) {
				t.Error("returned type does not deep equal to expected")
			}
		})
	}
}
