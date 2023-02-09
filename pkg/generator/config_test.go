package generator_test

import (
	"testing"

	"github.com/dnitsch/configmanager/internal/testutils"
	"github.com/dnitsch/configmanager/pkg/generator"
)

func TestParseTokens(t *testing.T) {
	ttests := map[string]struct {
		config   generator.GenVarsConfig
		rawToken string
		expect   generator.TokenConfigVars
	}{
		"role, version specified": {
			*generator.NewConfig(),
			`FOO://basjh/dskjuds/123|d88[role:arn:aws:iam::1111111:role/i-orchestration,version:1082313]`,
			generator.TokenConfigVars{
				Token:   `FOO://basjh/dskjuds/123|d88`,
				Role:    `arn:aws:iam::1111111:role/i-orchestration`,
				Version: `1082313`,
			},
		},
		"version, role specified": {
			*generator.NewConfig(),
			`FOO://basjh/dskjuds/123|d88[version:1082313,role:arn:aws:iam::1111111:role/i-orchestration]`,
			generator.TokenConfigVars{
				Token:   `FOO://basjh/dskjuds/123|d88`,
				Role:    `arn:aws:iam::1111111:role/i-orchestration`,
				Version: `1082313`,
			},
		},
		"version only specified": {
			*generator.NewConfig(),
			`FOO://basjh/dskjuds/123|d88[version:1082313]`,
			generator.TokenConfigVars{
				Token:   `FOO://basjh/dskjuds/123|d88`,
				Role:    ``,
				Version: `1082313`,
			},
		},
		"role only specified": {
			*generator.NewConfig(),
			`FOO://basjh/dskjuds/123|d88[role:arn:aws:iam::1111111:role/i-orchestration]`,
			generator.TokenConfigVars{
				Token:   `FOO://basjh/dskjuds/123|d88`,
				Role:    `arn:aws:iam::1111111:role/i-orchestration`,
				Version: ``,
			},
		},
		"no additional config specified": {
			*generator.NewConfig(),
			`FOO://basjh/dskjuds/123|d88`,
			generator.TokenConfigVars{
				Token:   `FOO://basjh/dskjuds/123|d88`,
				Role:    ``,
				Version: ``,
			},
		},
		"additional config specified but empty": {
			*generator.NewConfig(),
			`FOO://basjh/dskjuds/123`,
			generator.TokenConfigVars{
				Token:   `FOO://basjh/dskjuds/123`,
				Role:    ``,
				Version: ``,
			},
		},
	}
	for name, tt := range ttests {
		t.Run(name, func(t *testing.T) {
			got := tt.config.ParseTokenVars(tt.rawToken)
			if got.Role != tt.expect.Role {
				t.Errorf(testutils.TestPhraseWithContext, "role incorrect", got.Role, tt.expect.Role)
			}
			if got.Version != tt.expect.Version {
				t.Errorf(testutils.TestPhraseWithContext, "version incorrect", got.Version, tt.expect.Version)
			}
			if got.Token != tt.expect.Token {
				t.Errorf(testutils.TestPhraseWithContext, "token incorrect", got.Token, tt.expect.Token)
			}
		})
	}
}
