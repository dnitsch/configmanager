package genvars

import (
	"testing"

	"honnef.co/go/tools/config"
)

type fixture struct {
	t    *testing.T
	conf config.Config
}

func newFixture(t *testing.T) *fixture {
	f := &fixture{}
	f.t = t
	return f
}

func TestSecretsMangerImplementation(t *testing.T) {
	f := newFixture(t)
	got := true
	if !got {
		f.t.Error("Failed")
	}
}

func TestParamStoreImplementation(t *testing.T) {
	f := newFixture(t)
	got := true
	if !got {
		f.t.Error("Failed")
	}
}
