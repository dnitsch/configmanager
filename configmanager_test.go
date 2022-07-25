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

func (m *mockGenVars) ConvertToExportVar() {
}

func (m *mockGenVars) FlushToFile() (string, error) {
	return "", nil
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
