package config

import "testing"

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
			if tt.name != SELF_NAME {
				t.Error("self name does not match")
			}
		})
	}
}
