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
configmanager --tokens AWSSECRETS#/appxyz/service1-password --tokens AWSPARAMSTR#/appxyz/service1-password
source app.env
rm -f app.env
./startapp
```

By default the output path is `app.env` relative to the exec binary.

This can be overridden by passing in the `--path` param.

```bash
configmanager --token AWSSECRETS#/appxyz/service1-password --token AWSPARAMSTR#/appxyz/service12-settings --path /some/path/app.env
source /some/path/app.env
./startapp # psuedo script to start an application
```

Alternatively you can set the path as stdout which will reduce the need to save and source the env from file.

>!Warning! about eval - if you are retrieving secrets from sources you don't control the input of - best to stick wtih the file approach and then delete the file.

```bash
eval "$(configmanager r -t AWSSECRETS#/appxyz/service1-password -t AWSPARAMSTR#/appxyz/service12-settings -p stdout)" && ./.ignore-out.sh
```

The token is made up of 3 parts:

- `AWSSECRETS` the strategy identifier to choose at runtime

- `#` separator - used for

- `/path/to/parameter` the actual path to the secret or parameter in the target system e.g. AWS SecretsManager or ParameterStore (it does assume a path like pattern might throw a runtime error if not found)

If contents of the `AWSSECRETS#/appxyz/service1-password` are a string then `service1-password` will be the key and converted to UPPERCASE e.g. `SERVICE1_PASSWORD=som3V4lue`



## Go API

### Sample Use case

One of the sample use cases includes implementation in a K8s controller.

E.g. your Custom CRD stores some values in plain text that should really be secrets/nonpublic config parameters - something like this can be invoked from inside the controller code using the generator pkg API.

```go
func replaceTokens(in string, t *v1alpha.CustomFooCrdSpec) error {

	tokens := []string{}

	for k := range generator.VarPrefix {
		matches := regexp.MustCompile(`(?s)`+regexp.QuoteMeta(k)+`.([^\"]+)`).FindAllString(in, -1)
		tokens = append(tokens, matches...)
	}

	cnf := generator.GenVarsConfig{}
	m, err := configmanager.Retrieve(tokens, cnf)

	if err != nil {
		return err
	}

	if err := json.Unmarshal(replaceString(m, in), &t); err != nil {
		return err
	}
	return nil
}

func replaceString(inputMap generator.ParsedMap, inputString string) []byte {
	for oldVal, newVal := range inputMap {
		inputString = strings.ReplaceAll(inputString, oldVal, fmt.Sprint(newVal))
	}
	return []byte(inputString)
}
```

```yaml
apiVersion: crd.foo.custom/v1alpha1
kind: CustomFooCrd
metadata:
  name: foo
  namespace: bar
spec:
  name: baz
  secret_val: AWSSECRETS#/customfoo/secret-val
  owner: test_10016@example.com
```

Above example would ensure that you can safely store config/secret values on a CRD in plain text.

> Beware logging out the CRD after tokens have been replaced.

## Help

- More implementations should be easily added with a specific implementation under the strategy interface 
    - e.g. AzureKMS or GCP equivalent

- maybe run as cron in the background to perform a periodic sync in case values change?
