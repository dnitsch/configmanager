package configmanager

import (
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

func (m *mockGenVars) FlushToFile() error {
	return nil
}

func (m *mockGenVars) StrToFile(str string) error {
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
export FOO=FOO#/test
export FOO1=FOO#/test
export FOO2=FOO#/test
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
export FOO=val1
export FOO1=val1
export FOO2=val1
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
