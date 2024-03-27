package cmdutils_test

// func Test_ConvertToExportVars(t *testing.T) {
// 	tests := map[string]struct {
// 		rawMap       ParsedMap
// 		expectStr    string
// 		expectLength int
// 	}{
// 		"number included":     {ParsedMap{"foo": "BAR", "num": 123}, `export FOO='BAR'`, 2},
// 		"strings only":        {ParsedMap{"foo": "BAR", "num": "a123"}, `export FOO='BAR'`, 2},
// 		"numbers only":        {ParsedMap{"foo": 123, "num": 456}, `export FOO=123`, 2},
// 		"map inside response": {ParsedMap{"map": `{"foo":"bar","baz":"qux"}`, "num": 123}, `export FOO='bar'`, 3},
// 	}

// 	for name, tt := range tests {
// 		t.Run(name, func(t *testing.T) {
// 			f := newFixture(t)
// 			f.configGenVars(standardop, standardts)
// 			f.c.rawMap = muRawMap{tokenMap: tt.rawMap}
// 			f.c.ConvertToExportVar()
// 			got := f.c.outString
// 			if got == nil {
// 				t.Errorf(testutils.TestPhrase, got, "not nil")
// 			}
// 			if len(got) != tt.expectLength {
// 				t.Errorf(testutils.TestPhrase, len(got), tt.expectLength)
// 			}
// 			st := strings.Join(got, "\n")
// 			if !strings.Contains(st, tt.expectStr) {
// 				t.Errorf(testutils.TestPhrase, st, tt.expectStr)
// 			}
// 		})
// 	}
// }

// func Test_listToString(t *testing.T) {
// 	tests := map[string]struct {
// 		in     []string
// 		expect string
// 	}{
// 		"1 item slice": {[]string{"export ONE=foo"}, "export ONE=foo"},
// 		"0 item slice": {[]string{}, ""},
// 		"4 item slice": {[]string{"123", "123", "123", "123"}, `123
// 123
// 123
// 123`,
// 		},
// 	}
// 	for name, tt := range tests {
// 		t.Run(name, func(t *testing.T) {
// 			got := listToString(tt.in)
// 			if got != tt.expect {
// 				t.Errorf(testutils.TestPhrase, tt.expect, got)
// 			}
// 		})
// 	}
// }
