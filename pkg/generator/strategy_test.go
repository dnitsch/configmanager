package generator

import (
	"testing"
)

func Test_rsRetrieve(t *testing.T) {
	ttests := map[string]struct {
		impl       genVarsStrategy
		config     GenVarsConfig
		token      string
		implPrefix ImplementationPrefix
		expect     string
	}{
		"AWS success secrets manager ": {},
	}
	for name, tt := range ttests {
		t.Run(name, func(t *testing.T) {
			_ = tt.impl
		})
	}
}
