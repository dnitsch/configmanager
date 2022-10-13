# Config Manager

[![Go Report Card](https://goreportcard.com/badge/github.com/dnitsch/configmanager)](https://goreportcard.com/report/github.com/dnitsch/configmanager)

Package used for retrieving application settings from various sources.

Currently supported variable and secrets implementations:

- AWS SecretsManager
- AWS ParameterStore
- AzureKeyvault Secrets
- TODO:
  - GCP
  - Hashicorp  

The main driver is to use component level configuration objects, if stored in a `"namespaced"` manner e.g. in AWS ParamStore as `/nonprod/component-service-a/configVar`, however this is not a requirement and the param name can be whatever. Though whilst using some sort of a organised manner it will be more straight forward to allow other services to consume certain secrets/params based on resource/access policies. 

> Beware size limitation with certain config/vault implementations. In which case it's best to split certain items up e.g. TLS certs `/nonprod/component-service-a/pub-cert`, `/nonprod/component-service-a/private-cert`, `/nonprod/component-service-a/chain1-cert`, etc... 

Where `configVar` can be either a primitive type like a string `'som3#!S$CRet'` or a number `3306` or a parseable single level JSON object like `{host: ..., pass: ...., port: ...}` which can be returned whole or accessed via a key separator for a specific value.

## Use cases

- Kubernetes

   Avoid storing overly large configmaps and especially using secrets objects to store actual secrets e.g. DB passwords, 3rd party API creds, etc... By only storing a config file or a script containing only the tokens e.g. `AWSSECRETS#/$ENV/service/db-config` it can be git committed without writing numerous shell scripts, only storing either some interpolation vars like `$ENV` in a configmap or the entire configmanager token for smaller use cases.
- VMs
   VM deployments can function in a similar manner by passing in the contents or a path to the source config and the output path so that app at startup time can consume it.
- Functions (written in Go)
   Only storing tokens in env variables available to the function as plain text tokens gets around needing to store actual secrets in function env vars and can also be used across a variety of config stores.

## CLI

ConfigManager comes packaged as a CLI for all major platforms, to see [download/installation](./docs/installation.md)

For more detailed usage you can run -h with each subcommand and additional info can be found [here](./docs/commands.md)

## Config Tokens

The token is made up of 3 parts:

- `AWSSECRETS` the strategy identifier to choose at runtime

- `#` separator - used for separating the implementation indicator and the look up value.

> The default is currently `#` - it will change to `://` to allow for a more natural reading of the "token". you can achieve this behaviour now by either specifying the `-s` to the CLI or 
```go

rawStr := `somePAss: AWSPARAMSTR:///int-test/pocketbase/admin-pwd`
cm := configmanager.ConfigManager{}
// use custom token separator
// inline with v2 coming changes
cnf := generator.NewConfig().WithTokenSeparator("://")
replaced, err := cm.RetrieveWithInputReplaced(rawStr, *cnf)

```

- `/path/to/parameter` the actual path to the secret or parameter in the target system e.g. AWS SecretsManager or ParameterStore (it does assume a path like pattern might throw a runtime error if not found)

If contents of the `AWSSECRETS#/appxyz/service1-password` are a string then `service1-password` will be the key and converted to UPPERCASE e.g. `SERVICE1_PASSWORD=som3V4lue`

### KeySeparator

Specifying a key seperator on token items that can be parsed as a K/V map will result in only retrieving the specific key from the map. 

e.g. if contents of the `AWSSECRETS#/appxyz/service1-db-config` are parseable into the below object
```json
{
  "host": "db.internal",
  "port": 3306,
  "pass": "sUp3$ecr3T!",
}
```

Then you can access the single values like this `AWSSECRETS#/appxyz/service1-db-config|host` ==> `export SERVICE1_DB_CONFIG__HOST='db.internal'`

Alternatively if you are `configmanager`-ing a file via the fromstr command and the input is something like this:

(YAML)

```yaml
app:
  name: xyz
db:
  host: AWSSECRETS#/appxyz/service1-db-config|host
  port: AWSSECRETS#/appxyz/service1-db-config|port
  pass: AWSSECRETS#/appxyz/service1-db-config|pass
```

which would result in this

```yaml
app:
  name: xyz
db:
  host: db.internal
  port: 3306
  pass: sUp3$ecr3T!
```

If your config parameter matches the config interface, you can also leave the entire token to point to the `db` key


```yaml
app:
  name: xyz
db: AWSSECRETS#/appxyz/service1-db-config
```

result:

```yaml
app:
  name: xyz
db: {
  "host": "db.internal",
  "port": 3306,
  "pass": "sUp3$ecr3T!",
}
```

### Special consideration for AZKVSECRET

For Azure KeyVault the first part of the token needs to be the name of the vault.

> Azure Go SDK (v2) requires the vault Uri on initializing the client

`AZKVSECRET#/test-vault//token/1` ==> will use KeyVault implementation to retrieve the `/token/1` from a `test-vault`.

`AZKVSECRET#/test-vault/no-slash-token-1` ==> will use KeyVault implementation to retrieve the `no-slash-token-1` from a `test-vault`.

> The preceeding slash to the vault name is optional - `AZKVSECRET#/test-vault/no-slash-token-1` and `AZKVSECRET#test-vault/no-slash-token-1` will both identify the vault of name `test-vault`

## Go API

latest api [here](https://pkg.go.dev/github.com/dnitsch/configmanager)

### Sample Use case

One of the sample use cases includes implementation in a K8s controller.

E.g. your Custom CRD stores some values in plain text that should really be secrets/nonpublic config parameters - something like this can be invoked from inside the controller code using the generator pkg API.

See [examples](./examples/examples.go) for more examples and tests for sample input/usage

```go
package main

import (
	"context"
	"fmt"

	"github.com/dnitsch/configmanager/pkg/generator"
	"github.com/dnitsch/configmanager"
)

func main() {
	cm := &configmanager.ConfigManager{}
	cnf := generator.NewConfig()
	// JSON Marshal K8s CRD into
	exampleK8sCrdMarshalled := `apiVersion: crd.foo.custom/v1alpha1
kind: CustomFooCrd
metadata:
	name: foo
	namespace: bar
spec:
	name: baz
	secret_val: AWSSECRETS#/customfoo/secret-val
	owner: test_10016@example.com
`
	pm, err := cm.RetrieveWithInputReplaced(exampleK8sCrdMarshalled, *cnf)

	if err != nil {
		panic(err)
	}
	fmt.Println(pm)
}
```

Above example would ensure that you can safely store config/secret values on a CRD in plain text.

Or using go1.19+ [generics example](https://github.com/dnitsch/reststrategy/blob/d14ccec2b29bff646678ab9cf1775c0e93308569/controller/controller.go#L353).

> Beware logging out the CRD after tokens have been replaced.

Samlpe call to retrieve from inside an app/serverless function to only grab the relevant values from config.

```go
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/dnitsch/configmanager"
	"github.com/dnitsch/configmanager/pkg/generator"
)

var (
	DB_CONNECTION_STRING    string = "someuser:%v@tcp(%s:3306)/someschema"
	DB_PASSWORD_SECRET_PATH string = os.Getenv("DB_PASSWORD_TOKEN")
	DB_HOST_URL             string = os.Getenv("DB_URL_TOKEN")
)

func main() {
	connString, err := credentialString(context.TODO, DB_PASSWORD_SECRET_PATH, DB_HOST_URL)
	if err != nil {
		log.Fatal(err)
	}

}

func credentialString(ctx context.Context, pwdToken, hostToken string) (string, error) {

	cnf := generator.NewConfig()

	pm, err := configmanager.Retrieve([]string{pwdToken, hostToken}, *cnf)

	if err != nil {
		return "", err
	}
	if pwd, ok := pm[pwdToken]; ok {
		if host, ok := pm[hostToken]; ok {
			return fmt.Sprintf(DB_CONNECTION_STRING, pwd, host), nil
		}
	}

	return "", fmt.Errorf("unable to find value via token")
}
```

## Help

- More implementations should be easily added with a specific implementation under the strategy interface
    - e.g. GCP equivalent

- maybe run as cron in the background to perform a periodic sync in case values change?
