package utils

import "testing"

func Test_generateFromStr(t *testing.T) {
	tests := []struct {
		name string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			// TODO: finish off tests
			if err := generateStrOutFromInput(nil, nil, ""); err != nil {
				t.Error(err)
			}
		})
	}
}
