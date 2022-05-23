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

### TODO

- Go API and GoDoc 

- Ideally should be split into numerous packages for reading vars and creating using the same token reversal

## Usage

```bash
configmanager --tokens AWSSECRETS#/appxyz/service1-password --tokens AWSPARAMSTR#/appxyz/service1-password
source app.env
./startapp
```

By default the output path is `app.env` relative to the exec binary.

This can be overridden by passing in the `--path` param.

```bash
configmanager --token AWSSECRETS#/appxyz/service1-password --token AWSPARAMSTR#/appxyz/service12-settings --path /some/path/app.env
source /some/path/app.env
./startapp # psuedo script to start an application
```

The token is made up of 3 parts:

- `AWSSECRETS` the strategy identifier to choose at runtime

- `#` separator - used for

- `/path/to/parameter` the actual path to the secret or parameter in the target system e.g. AWS SecretsManager or ParameterStore (it does assume a path like pattern might throw a runtime error if not found)

If contents of the `AWSSECRETS#/appxyz/service1-password` are a string then `service1-password` will be the key and converted to UPPERCASE e.g. `SERVICE1_PASSWORD=som3V4lue`

## Help

- More implementations should be easily added with a specific implementation under the strategy interface

- maybe run as cron in the background to perform a periodic sync in case values change?

TODO: lots more documentation and UNITTESTS
