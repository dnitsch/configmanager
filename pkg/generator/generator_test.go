package generator

import (
	"fmt"
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
	t *testing.T
	c *GenVars
}

func newFixture(t *testing.T) *fixture {
	f := &fixture{}
	f.t = t
	return f
}

func (f *fixture) goodGenVars(op, ts string) {
	conf := NewConfig().WithOutputPath(op).WithTokenSeparator(ts)
	gv := NewGenerator().WithConfig(conf)
	f.c = gv
}

func TestGenVarsWithConfig(t *testing.T) {

	f := newFixture(t)

	f.goodGenVars(customop, customts)
	if f.c.config.outpath != customop {
		f.t.Errorf(testutils.TestPhrase, customop, f.c.config.outpath)
	}
	if f.c.config.tokenSeparator != customts {
		f.t.Errorf(testutils.TestPhrase, customts, f.c.config.tokenSeparator)
	}
}

func TestStripPrefixNormal(t *testing.T) {

	want := "/normal/without/prefix"
	prefix := SecretMgrPrefix
	f := newFixture(t)
	f.goodGenVars(standardop, standardts)

	got := f.c.stripPrefix(fmt.Sprintf("%s#%s", prefix, want), prefix)
	if got != want {
		f.t.Errorf(testutils.TestPhrase, want, got)
	}

	gotNegative := f.c.stripPrefix(fmt.Sprintf("%s___%s", prefix, want), prefix)
	if gotNegative == want {
		f.t.Errorf(testutils.TestPhrase, want, gotNegative)
	}
}

func TestNormalizedMapWithString(t *testing.T) {
	a := map[string]any{"foo": "bar"}
	got := envVarNormalize(a)
	for k := range got {
		if k != "FOO" {
			t.Errorf(testutils.TestPhrase, "FOO", k)
		}
	}
}

func TestNormalizedMapWithInt(t *testing.T) {
	a := map[string]any{"num": 123}
	got := envVarNormalize(a)
	for k := range got {
		if k != "NUM" {
			t.Errorf(testutils.TestPhrase, "NUM", k)
		}
	}
}

func Test_ConvertToExportVars(t *testing.T) {
	tests := []struct {
		name   string
		rawMap ParsedMap
		expect []string
	}{
		{
			name:   "number included",
			rawMap: ParsedMap{"foo": "BAR", "num": 123},
			expect: []string{`export FOO='BAR'`, `export NUM=123`},
		},
		{
			name:   "strings only",
			rawMap: ParsedMap{"foo": "BAR", "num": "a123"},
			expect: []string{`export FOO='BAR'`, `export NUM='a123'`},
		},
		{
			name:   "numbers only",
			rawMap: ParsedMap{"foo": 123, "num": 456},
			expect: []string{`export FOO=123`, `export NUM=456`},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newFixture(t)
			f.goodGenVars(standardop, standardts)
			f.c.rawMap = tt.rawMap
			f.c.ConvertToExportVar()
			got := f.c.outString
			if got == nil {
				t.Errorf(testutils.TestPhrase, "not nil", got)
			}
			if len(tt.expect) != len(got) {
				t.Errorf(testutils.TestPhrase, len(tt.expect), len(got))

			}
			for k, v := range got {
				if v != tt.expect[k] {
					t.Errorf(testutils.TestPhrase, tt.expect[k], got[k])
				}
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
