package generator_test

// func TestParseTokens(t *testing.T) {
// 	ttests := map[string]struct {
// 		config   generator.GenVarsConfig
// 		rawToken string
// 		expect   generator.TokenConfigVars
// 	}{
// 		"role, version specified": {
// 			*generator.NewConfig(),
// 			`FOO://basjh/dskjuds/123|d88[role:arn:aws:iam::1111111:role/i-orchestration,version:1082313]`,
// 			generator.TokenConfigVars{
// 				Token:   `FOO://basjh/dskjuds/123|d88`,
// 				Role:    `arn:aws:iam::1111111:role/i-orchestration`,
// 				Version: `1082313`,
// 			},
// 		},
// 		"version, role specified": {
// 			*generator.NewConfig(),
// 			`FOO://basjh/dskjuds/123|d88[version:1082313,role:arn:aws:iam::1111111:role/i-orchestration]`,
// 			generator.TokenConfigVars{
// 				Token:   `FOO://basjh/dskjuds/123|d88`,
// 				Role:    `arn:aws:iam::1111111:role/i-orchestration`,
// 				Version: `1082313`,
// 			},
// 		},
// 		"version only specified": {
// 			*generator.NewConfig(),
// 			`FOO://basjh/dskjuds/123|d88[version:1082313]`,
// 			generator.TokenConfigVars{
// 				Token:   `FOO://basjh/dskjuds/123|d88`,
// 				Role:    ``,
// 				Version: `1082313`,
// 			},
// 		},
// 		"role only specified": {
// 			*generator.NewConfig(),
// 			`FOO://basjh/dskjuds/123|d88[role:arn:aws:iam::1111111:role/i-orchestration]`,
// 			generator.TokenConfigVars{
// 				Token:   `FOO://basjh/dskjuds/123|d88`,
// 				Role:    `arn:aws:iam::1111111:role/i-orchestration`,
// 				Version: ``,
// 			},
// 		},
// 		"no additional config specified": {
// 			*generator.NewConfig(),
// 			`FOO://basjh/dskjuds/123|d88`,
// 			generator.TokenConfigVars{
// 				Token:   `FOO://basjh/dskjuds/123|d88`,
// 				Role:    ``,
// 				Version: ``,
// 			},
// 		},
// 		"additional config specified but empty": {
// 			*generator.NewConfig(),
// 			`FOO://basjh/dskjuds/123[]`,
// 			generator.TokenConfigVars{
// 				Token:   `FOO://basjh/dskjuds/123`,
// 				Role:    ``,
// 				Version: ``,
// 			},
// 		},
// 	}
// 	for name, tt := range ttests {
// 		t.Run(name, func(t *testing.T) {
// 			got := tt.config.ParseTokenVars(tt.rawToken)
// 			if got.Role != tt.expect.Role {
// 				t.Errorf(testutils.TestPhraseWithContext, "role incorrect", got.Role, tt.expect.Role)
// 			}
// 			if got.Version != tt.expect.Version {
// 				t.Errorf(testutils.TestPhraseWithContext, "version incorrect", got.Version, tt.expect.Version)
// 			}
// 			if got.Token != tt.expect.Token {
// 				t.Errorf(testutils.TestPhraseWithContext, "token incorrect", got.Token, tt.expect.Token)
// 			}
// 		})
// 	}
// }

// func Test_MarshalMetadata_with_label_struct_succeeds(t *testing.T) {
// 	type labelMeta struct {
// 		Label string `json:"label"`
// 	}

// 	ttests := map[string]struct {
// 		config   *generator.GenVarsConfig
// 		rawToken string
// 		wantVal  any
// 	}{
// 		"when provider expects label on token and label exists": {
// 			generator.NewConfig(),
// 			`FOO://basjh/dskjuds/123|d88[label=dev]`,
// 			"dev",
// 		},
// 		"when provider expects label on token and label does not exist": {
// 			generator.NewConfig(),
// 			`FOO://basjh/dskjuds/123|d88[someother=dev]`,
// 			"",
// 		},
// 		"no metadata found": {
// 			generator.NewConfig(),
// 			`FOO://basjh/dskjuds/123|d88`,
// 			"",
// 		},
// 		"no metadata found incorrect marker placement": {
// 			generator.NewConfig(),
// 			`FOO://basjh/dskjuds/123|d88]asdas=bar[`,
// 			"",
// 		},
// 	}
// 	for name, tt := range ttests {
// 		t.Run(name, func(t *testing.T) {
// 			inputTyp := &labelMeta{}
// 			got := tt.config.ParseMetadata(tt.rawToken, inputTyp)
// 			if got == nil {
// 				t.Errorf("got <nil>, wanted: %v\n", inputTyp)
// 			}

// 			gotTyp := got.(*labelMeta)

// 			if gotTyp.Label != tt.wantVal {
// 				t.Errorf("got %v, wanted: %v\n", gotTyp, inputTyp)
// 			}
// 		})
// 	}
// }

// func Test_MarshalMetadata_with_label_struct_fails_with_nil_pointer(t *testing.T) {
// 	got := generator.NewConfig().ParseMetadata(`FOO://basjh/dskjuds/123|d88`, nil)
// 	if got != nil {
// 		t.Errorf("got %v, wanted:<nil>\n", got)
// 	}

// }
