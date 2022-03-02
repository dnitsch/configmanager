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
	t    *testing.T
	c    *genVars
	conf GenVarsConfig
}

func newFixture(t *testing.T) *fixture {
	f := &fixture{}
	f.t = t
	return f
}

func (f *fixture) goodGenVars(op, ts string) {
	f.conf = GenVarsConfig{Outpath: op, TokenSeparator: ts}

	gv := New()
	gv.WithConfig(&f.conf)
	f.c = gv
}

func TestGenVarsWithConfig(t *testing.T) {

	f := newFixture(t)

	f.goodGenVars(customop, customts)
	if f.conf.Outpath != customop {
		f.t.Errorf(testutils.TestPhrase, customop, f.conf.Outpath)
	}
	if f.conf.TokenSeparator != customts {
		f.t.Errorf(testutils.TestPhrase, customts, f.conf.TokenSeparator)
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
	a := map[string]interface{}{"foo": "bar"}
	got := envVarNormalize(a)
	for k := range got {
		if k != "FOO" {
			t.Errorf(testutils.TestPhrase, "FOO", k)
		}
	}
}

func TestNormalizedMapWithInt(t *testing.T) {
	a := map[string]interface{}{"num": 123}
	got := envVarNormalize(a)
	for k := range got {
		if k != "NUM" {
			t.Errorf(testutils.TestPhrase, "NUM", k)
		}
	}
}

func TestConvertToExportVars(t *testing.T) {
	want := `export FOO='BAR'
export NUM=123
`
	m := ParsedMap{}
	m["foo"] = "BAR"
	m["num"] = 123
	f := newFixture(t)
	f.goodGenVars(standardop, standardts)
	f.c.rawMap = m
	f.c.ConvertToExportVar()
	got := f.c.outString
	if got != want {
		t.Errorf(testutils.TestPhrase, want, got)
	}
}
