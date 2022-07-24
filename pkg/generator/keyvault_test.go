package generator

import (
	"testing"

	"github.com/dnitsch/configmanager/internal/testutils"
)

func Test_azSplitToken(t *testing.T) {
	tests := []struct {
		name   string
		token  string
		expect azVaultHelper
	}{
		{
			name:  "simple_with_preceding_slash",
			token: "/test-vault/somejsontest",
			expect: azVaultHelper{
				vaultUri: "https://test-vault.vault.azure.net",
				token:    "somejsontest",
			},
		},
		{
			name:  "missing_initial_slash",
			token: "test-vault/somejsontest",
			expect: azVaultHelper{
				vaultUri: "https://test-vault.vault.azure.net",
				token:    "somejsontest",
			},
		},
		{
			name:  "missing_initial_slash_multislash_secretname",
			token: "test-vault/some/json/test",
			expect: azVaultHelper{
				vaultUri: "https://test-vault.vault.azure.net",
				token:    "some/json/test",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := azSplitToken(tt.token)
			if got.token != tt.expect.token {
				t.Errorf(testutils.TestPhrase, tt.expect.token, got.token)
			}
			if got.vaultUri != tt.expect.vaultUri {
				t.Errorf(testutils.TestPhrase, tt.expect.vaultUri, got.vaultUri)
			}
		})
	}
}
