# adding provider

Add Token 

`VarPrefix = map[string]bool{SecretMgrPrefix: true, ParamStorePrefix: true, AzKeyVaultSecretsPrefix: true, GcpSecretMgrPrefix: true}` // <-- ADD here

```go
const (
	// tokenSeparator used for identifying the end of a prefix and beginning of token
	// see notes about special consideration for AZKVSECRET tokens
	tokenSeparator = "#"
	// keySeparator used for accessing nested objects within the retrieved map
	keySeparator = "|"
	// AWS SecretsManager prefix
	SecretMgrPrefix = "AWSSECRETS"
	// AWS Parameter Store prefix
	ParamStorePrefix = "AWSPARAMSTR"
	// Azure Key Vault Secrets prefix
	AzKeyVaultSecretsPrefix = "AZKVSECRET"
	// GCP SecretsManager prefix
	GcpSecretMgrPrefix = "GCPSECRETS" // <-- ADD here
)
```

inside 

```go
func (imp *GcpSecrets) getTokenValue(v *retrieveStrategy) (string, error) {

	log.Infof("%s", "Concrete implementation GcpSecrets")
	log.Infof("Getting Secret: %s", imp.token)

	input := &gcpsecretspb.AccessSecretVersionRequest{
		Name: fmt.Sprintf("%s/versions/latest", v.stripPrefix(imp.token, GcpSecretsPrefix)), // <-- Ensure this is set correctly
	}
```