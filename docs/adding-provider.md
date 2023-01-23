# adding provider

Add Token Prefix

```go
const (
	// AWS SecretsManager prefix
	SecretMgrPrefix ImplementationPrefix = "AWSSECRETS"
	// AWS Parameter Store prefix
	ParamStorePrefix ImplementationPrefix = "AWSPARAMSTR"
	// Azure Key Vault Secrets prefix
	AzKeyVaultSecretsPrefix ImplementationPrefix = "AZKVSECRET"
	// Hashicorp Vault prefix
	HashicorpVaultPrefix ImplementationPrefix = "VAULT"
	// GcpSecrets
	GcpSecretsPrefix ImplementationPrefix = "GCPSECRETS"
)
```

```go
var (
	// default varPrefix used by the replacer function
	// any token must beging with one of these else
	// it will be skipped as not a replaceable token
	VarPrefix = map[ImplementationPrefix]bool{SecretMgrPrefix: true, ParamStorePrefix: true, AzKeyVaultSecretsPrefix: true, GcpSecretsPrefix: true, HashicorpVaultPrefix: true} // <-- ADD here
)
```

ensure your implementation satisfy the `genVarsStrategy` interface

```go
type genVarsStrategy interface {
	getTokenValue(rs *retrieveStrategy) (s string, e error)
	setToken(s string)
	setValue(s string)
}
```

Even if the native type is K/V return a marshalled version of the JSON as the rest of the flow will decide how to present it back to the final consumer.
