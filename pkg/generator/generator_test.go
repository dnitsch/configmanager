package generator

import (
	"fmt"
	"strings"
	"testing"

	"github.com/dnitsch/configmanager/internal/testutils"
)

var (
	customts   = "___"
	customop   = "/foo"
	standardop = "./app.env"
	standardts = "#"
)

type fixture struct {
	t  *testing.T
	c  *GenVars
	rs *retrieveStrategy
}

func newFixture(t *testing.T) *fixture {
	f := &fixture{}
	f.t = t
	return f
}

func (f *fixture) configGenVars(op, ts string) {
	conf := NewConfig().WithOutputPath(op).WithTokenSeparator(ts)
	gv := NewGenerator().WithConfig(conf)
	f.rs = newRetrieveStrategy(NewDefatultStrategy(), *conf)
	f.c = gv
}

func TestGenVarsWithConfig(t *testing.T) {

	f := newFixture(t)

	f.configGenVars(customop, customts)
	if f.c.config.outpath != customop {
		f.t.Errorf(testutils.TestPhrase, f.c.config.outpath, customop)
	}
	if f.c.config.tokenSeparator != customts {
		f.t.Errorf(testutils.TestPhrase, f.c.config.tokenSeparator, customts)
	}
}

func TestStripPrefixNormal(t *testing.T) {
	ttests := map[string]struct {
		prefix         ImplementationPrefix
		token          string
		keySeparator   string
		tokenSeparator string
		f              *fixture
		expect         string
	}{
		"standard azkv":               {AzKeyVaultSecretsPrefix, "AZKVSECRET://vault1/secret2", "|", "://", newFixture(t), "vault1/secret2"},
		"standard hashivault":         {HashicorpVaultPrefix, "VAULT://vault1/secret2", "|", "://", newFixture(t), "vault1/secret2"},
		"custom separator hashivault": {HashicorpVaultPrefix, "VAULT#vault1/secret2", "|", "#", newFixture(t), "vault1/secret2"},
	}
	for name, tt := range ttests {
		t.Run(name, func(t *testing.T) {
			tt.f.configGenVars(tt.keySeparator, tt.tokenSeparator)
			got := tt.f.rs.stripPrefix(tt.token, tt.prefix)
			if got != tt.expect {
				t.Errorf(testutils.TestPhrase, got, tt.expect)
			}
		})
	}
}

func Test_stripPrefix(t *testing.T) {
	f := newFixture(t)
	f.configGenVars(standardop, standardts)
	tests := []struct {
		name   string
		token  string
		prefix ImplementationPrefix
		expect string
	}{
		{
			name:   "simple",
			token:  fmt.Sprintf("%s#/test/123", SecretMgrPrefix),
			prefix: SecretMgrPrefix,
			expect: "/test/123",
		},
		{
			name:   "key appended",
			token:  fmt.Sprintf("%s#/test/123|key", ParamStorePrefix),
			prefix: ParamStorePrefix,
			expect: "/test/123",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := f.rs.stripPrefix(tt.token, tt.prefix)
			if tt.expect != got {
				t.Errorf(testutils.TestPhrase, tt.expect, got)
			}
		})
	}
}

func Test_NormaliseMap(t *testing.T) {
	f := newFixture(t)
	f.configGenVars(standardop, standardts)
	tests := []struct {
		name     string
		gv       *GenVars
		input    map[string]any
		expected string
	}{
		{
			name:     "foo->FOO",
			gv:       f.c,
			input:    map[string]any{"foo": "bar"},
			expected: "FOO",
		},
		{
			name:     "num->NUM",
			gv:       f.c,
			input:    map[string]any{"num": 123},
			expected: "NUM",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := f.c.envVarNormalize(tt.input)
			for k := range got {
				if k != tt.expected {
					t.Errorf(testutils.TestPhrase, tt.expected, k)
				}
			}
		})
	}
}

func Test_KeyLookup(t *testing.T) {
	f := newFixture(t)
	f.configGenVars(standardop, standardts)

	tests := []struct {
		name   string
		gv     *GenVars
		val    string
		key    string
		expect string
	}{
		{
			name:   "lowercase key found in str val",
			gv:     f.c,
			key:    `something|key`,
			val:    `{"key": "11235"}`,
			expect: "11235",
		},
		{
			name:   "lowercase key found in numeric val",
			gv:     f.c,
			key:    `something|key`,
			val:    `{"key": 11235}`,
			expect: "11235",
		},
		{
			name:   "uppercase key found in val",
			gv:     f.c,
			key:    `something|KEY`,
			val:    `{"KEY": "upposeres"}`,
			expect: "upposeres",
		},
		{
			name:   "no key found in val",
			gv:     f.c,
			key:    `something`,
			val:    `{"key": "notfound"}`,
			expect: `{"key": "notfound"}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := f.c.keySeparatorLookup(tt.key, tt.val)
			if got != tt.expect {
				t.Errorf(testutils.TestPhrase, tt.expect, got)
			}
		})
	}
}

func Test_ConvertToExportVars(t *testing.T) {
	tests := []struct {
		name   string
		rawMap ParsedMap
		expect string
	}{
		{
			name:   "number included",
			rawMap: ParsedMap{"foo": "BAR", "num": 123},
			expect: `export FOO='BAR'`,
		},
		{
			name:   "strings only",
			rawMap: ParsedMap{"foo": "BAR", "num": "a123"},
			expect: `export FOO='BAR'`,
		},
		{
			name:   "numbers only",
			rawMap: ParsedMap{"foo": 123, "num": 456},
			expect: `export FOO=123`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newFixture(t)
			f.configGenVars(standardop, standardts)
			f.c.rawMap = tt.rawMap
			f.c.ConvertToExportVar()
			got := f.c.outString
			if got == nil {
				t.Errorf(testutils.TestPhrase, "not nil", got)
			}
			if 2 != len(got) {
				t.Errorf(testutils.TestPhrase, 2, len(got))

			}
			st := strings.Join(got, "\n")
			if !strings.Contains(st, tt.expect) {
				t.Errorf(testutils.TestPhrase, tt.expect, st)
			}
		})
	}
}

func Test_listToString(t *testing.T) {
	tests := []struct {
		name   string
		in     []string
		expect string
	}{
		{
			name:   "1 item slice",
			in:     []string{"export ONE=foo"},
			expect: "export ONE=foo",
		},
		{
			name:   "0 item slice",
			in:     []string{},
			expect: "",
		},
		{
			name: "4 item slice",
			in:   []string{"123", "123", "123", "123"},
			expect: `123
123
123
123`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := listToString(tt.in)
			if got != tt.expect {
				t.Errorf(testutils.TestPhrase, tt.expect, got)
			}
		})
	}
}
