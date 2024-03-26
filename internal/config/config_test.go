package config_test

import (
	"testing"

	"github.com/dnitsch/configmanager/internal/config"
	"github.com/dnitsch/configmanager/internal/testutils"
)

func Test_SelfName(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "configmanager",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.name != config.SELF_NAME {
				t.Error("self name does not match")
			}
		})
	}
}

func Test_MarshalMetadata_with_label_struct_succeeds(t *testing.T) {
	type labelMeta struct {
		Label string `json:"label"`
	}

	ttests := map[string]struct {
		config                *config.GenVarsConfig
		rawToken              string
		wantLabel             string
		wantMetaStrippedToken string
	}{
		"when provider expects label on token and label exists": {
			config.NewConfig().WithTokenSeparator("://"),
			`AZTABLESTORE://basjh/dskjuds/123|d88[label=dev]`,
			"dev",
			"AZTABLESTORE://basjh/dskjuds/123|d88",
		},
		"when provider expects label on token and label does not exist": {
			config.NewConfig().WithTokenSeparator("://"),
			`AZTABLESTORE://basjh/dskjuds/123|d88[someother=dev]`,
			"",
			"AZTABLESTORE://basjh/dskjuds/123|d88",
		},
		"no metadata found": {
			config.NewConfig().WithTokenSeparator("://"),
			`AZTABLESTORE://basjh/dskjuds/123|d88`,
			"",
			"AZTABLESTORE://basjh/dskjuds/123|d88",
		},
		"no metadata found incorrect marker placement": {
			config.NewConfig().WithTokenSeparator("://"),
			`AZTABLESTORE://basjh/dskjuds/123|d88]asdas=bar[`,
			"",
			"AZTABLESTORE://basjh/dskjuds/123|d88]asdas=bar[",
		},
		"no metadata found incorrect marker placement and no key separator": {
			config.NewConfig().WithTokenSeparator("://"),
			`AZTABLESTORE://basjh/dskjuds/123]asdas=bar[`,
			"",
			"AZTABLESTORE://basjh/dskjuds/123]asdas=bar[",
		},
		"no end found incorrect marker placement and no key separator": {
			config.NewConfig().WithTokenSeparator("://"),
			`AZTABLESTORE://basjh/dskjuds/123[asdas=bar`,
			"",
			"AZTABLESTORE://basjh/dskjuds/123[asdas=bar",
		},
		"no start found incorrect marker placement and no key separator": {
			config.NewConfig().WithTokenSeparator("://"),
			`AZTABLESTORE://basjh/dskjuds/123]asdas=bar]`,
			"",
			"AZTABLESTORE://basjh/dskjuds/123]asdas=bar]",
		},
		"metadata is in the middle of path lookup": {
			config.NewConfig().WithTokenSeparator("://"),
			`AZTABLESTORE://basjh/dskjuds/123[label=bar]|lookup`,
			"bar",
			"AZTABLESTORE://basjh/dskjuds/123|lookup",
		},
	}
	for name, tt := range ttests {
		t.Run(name, func(t *testing.T) {
			inputTyp := &labelMeta{}
			got := config.NewParsedTokenConfig(tt.rawToken, *tt.config)

			if got == nil {
				t.Errorf(testutils.TestPhraseWithContext, "Unable to parse token", nil, config.ParsedTokenConfig{})
			}

			got.ParseMetadata(inputTyp)

			if got.StripMetadata() != tt.wantMetaStrippedToken {
				t.Errorf(testutils.TestPhraseWithContext, "Token does not match", got.StripMetadata(), tt.wantMetaStrippedToken)
			}

			if inputTyp.Label != tt.wantLabel {
				t.Errorf(testutils.TestPhraseWithContext, "Metadata Label does not match", inputTyp.Label, tt.wantLabel)
			}
		})
	}
}

func Test_StripPrefix(t *testing.T) {
	ttests := map[string]struct {
		objType any
	}{
		"test1": {
			objType: nil,
		},
	}
	for name, _ := range ttests {
		t.Run(name, func(t *testing.T) {

		})
	}
}
