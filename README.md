# Config Manager

[![Go Report Card](https://goreportcard.com/badge/github.com/dnitsch/configmanager)](https://goreportcard.com/report/github.com/dnitsch/configmanager)

Package used for retrieving application settings from various sources.

Currently supported variable and secrets implementations - AWS SecretsManager and AWS ParameterStore.

> The Intended use is within containers to generate required application config that can later be sourced and then read by application at start up time so that e.g. K8s secrets don't need to be used for true "secrets" or jsut to avoid overly large configmaps

Must reference an existing token in a path like format.

GenVars will then write them to a file in this format:

```bash
export VAR=VALUE
export VAR2=VALUE2
...
```

## Installation

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
  insert       Retrieves a value for token(s) specified and optionally writes to a file
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

Replaces all the occurences of tokens inside strings and writes them back out to files.



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

## Help

- More implementations should be easily added with a specific implementation under the strategy interface
    - e.g. AzureKMS or GCP equivalent

- maybe run as cron in the background to perform a periodic sync in case values change?
