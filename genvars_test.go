package genvars

import (
	"context"
	"fmt"
	"testing"

	"github.com/dnitsch/genvars/internal/testutils"
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

	gv := NewGenVars("foobar", context.TODO())
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
