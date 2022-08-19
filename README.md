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

## CLI Installation

Major platform binaries [here](https://github.com/dnitsch/configmanager/releases)

*nix binary

```bash
curl -L https://github.com/dnitsch/configmanager/releases/latest/download/configmanager-linux -o configmanager
```

MacOS binary

```bash
curl -L https://github.com/dnitsch/configmanager/releases/latest/download/configmanager-darwin -o configmanager
```

```bash
chmod +x configmanager
sudo mv configmanager /usr/local/bin
```

Download specific version:

```bash
curl -L https://github.com/dnitsch/configmanager/releases/download/v0.5.0/configmanager-`uname -s` -o configmanager
```

## Usage

```bash
configmanager CLI for retrieving config or secret variables.
                Using a specific tokens as an array item

Usage:
  configmanager [command]

Available Commands:
  completion   Generate the autocompletion script for the specified shell
  help         Help about any command
  insert       Not yet implemented
  retrieve     Retrieves a value for token(s) specified
  string-input Retrieves all found token values in a specified string input
  version      Get version number configmanager

Flags:
  -h, --help                     help for configmanager
  -k, --key-separator string     Separator to use to mark a key look up in a map. e.g. AWSSECRETS#/token/map|key1 (default "|")
  -s, --token-separator string   Separator to use to mark concrete store and the key within it (default "#")
  -v, --verbose                  Verbosity level
```

### Commands

#### retrieve

Useful for retrieving a series of tokens in CI or before app start

```bash
configmanager retrieve --token AWSSECRETS#/appxyz/service1-password --token AWSPARAMSTR#/appxyz/service2-password
source app.env
```

This will have written to a defaul out path `app.env` in current directory the below contents

```bash
export SERVICE1_PASSWORD='somepass!$@sdhf'
export SERVICE2_PASSWORD='somepa22$!$'
```

Once sourced you could delete the file, however the environment variables will persist in the process info `/proc/someprocess`

```bash
rm -f app.env
./startapp
```

By default the output path is `app.env` relative to the exec binary.

This can be overridden by passing in the `--path` param.

```bash
configmanager retrieve --token AWSSECRETS#/appxyz/service1-password --token AWSPARAMSTR#/appxyz/service12-settings --path /some/path/app.env
source /some/path/app.env
./startapp # psuedo script to start an application
```

Alternatively you can set the path as stdout which will reduce the need to save and source the env from file.

>!Warning! about eval - if you are retrieving secrets from sources you don't control the input of - best to stick wtih the file approach and then delete the file.

```bash
eval "$(configmanager r -t AWSSECRETS#/appxyz/service1-password -t AWSPARAMSTR#/appxyz/service12-settings -p stdout)" && ./.ignore-out.sh
```

#### string-input

Replaces all the occurences of tokens inside strings and writes them back out to a file provided. 

This method can be used with entire application property files such as `application.yml` or `application.properties` for springboot apps or netcore app config in which ever format.

The `fromstr` (alias for `string-input`) respects all indentations so can be used on contents of a file of any type



## Config Tokens

The token is made up of 3 parts:

- `AWSSECRETS` the strategy identifier to choose at runtime

- `#` separator - used for

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

### Special AZKVSECRETS

For Azure KeyVault the first part of the token needs to be the name of the vault.

> Azure Go SDK (v2) requires the vault Uri on initializing the client

`AZKVSECRET#/test-vault//token/1` ==> will use KeyVault implementation to retrieve the `/token/1` from a `test-vault`.

`AZKVSECRET#/test-vault/no-slash-token-1` ==> will use KeyVault implementation to retrieve the `no-slash-token-1` from a `test-vault`.

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

> Beware logging out the CRD after tokens have been replaced.

Samlpe call to retrieve from inside an app/serverless function.

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
