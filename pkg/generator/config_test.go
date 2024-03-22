package generator_test

import (
	"testing"

	"github.com/dnitsch/configmanager/internal/testutils"
	"github.com/dnitsch/configmanager/pkg/generator"
)

func Test_MarshalMetadata_with_label_struct_succeeds(t *testing.T) {
	type labelMeta struct {
		Label string `json:"label"`
	}

	ttests := map[string]struct {
		config    *generator.GenVarsConfig
		rawToken  string
		wantLabel string
		wantToken string
	}{
		"when provider expects label on token and label exists": {
			generator.NewConfig(),
			`FOO://basjh/dskjuds/123|d88[label=dev]`,
			"dev",
			"FOO://basjh/dskjuds/123|d88",
		},
		"when provider expects label on token and label does not exist": {
			generator.NewConfig(),
			`FOO://basjh/dskjuds/123|d88[someother=dev]`,
			"",
			"FOO://basjh/dskjuds/123|d88",
		},
		"no metadata found": {
			generator.NewConfig(),
			`FOO://basjh/dskjuds/123|d88`,
			"",
			"FOO://basjh/dskjuds/123|d88",
		},
		"no metadata found incorrect marker placement": {
			generator.NewConfig(),
			`FOO://basjh/dskjuds/123|d88]asdas=bar[`,
			"",
			"FOO://basjh/dskjuds/123|d88]asdas=bar[",
		},
		"no metadata found incorrect marker placement and no key separator": {
			generator.NewConfig(),
			`FOO://basjh/dskjuds/123]asdas=bar[`,
			"",
			"FOO://basjh/dskjuds/123]asdas=bar[",
		},
		"no end found incorrect marker placement and no key separator": {
			generator.NewConfig(),
			`FOO://basjh/dskjuds/123[asdas=bar`,
			"",
			"FOO://basjh/dskjuds/123[asdas=bar",
		},
		"no start found incorrect marker placement and no key separator": {
			generator.NewConfig(),
			`FOO://basjh/dskjuds/123]asdas=bar]`,
			"",
			"FOO://basjh/dskjuds/123]asdas=bar]",
		},
		"metadata is in the middle of path lookup": {
			generator.NewConfig(),
			`FOO://basjh/dskjuds/123[label=bar]|lookup`,
			"bar",
			"FOO://basjh/dskjuds/123|lookup",
		},
	}
	for name, tt := range ttests {
		t.Run(name, func(t *testing.T) {
			inputTyp := &labelMeta{}
			got := generator.ParseMetadata(tt.rawToken, inputTyp)

			if got != tt.wantToken {
				t.Errorf(testutils.TestPhraseWithContext, "Token does not match", got, tt.wantToken)
			}

			if inputTyp.Label != tt.wantLabel {
				t.Errorf(testutils.TestPhraseWithContext, "Metadata Label does not match", inputTyp.Label, tt.wantLabel)
			}
		})
	}
}
