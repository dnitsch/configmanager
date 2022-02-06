# GENVARS

Package used for retrieving application settings from various sources.

Currently supported variable and secrets implementations - AWS SecretsManager and AWS ParameterStore.

> The Intended use is within containers to generate required application config that can later be sourced and then read by application 

Must reference an existing token in a path like format.

GenVars will then write them to a file in this format: 

```bash
export VAR=VALUE
export VAR2=VALUE2
...
```

## Usage

```bash
genvars --tokens AWSSECRETS#/appxyz/service1-password --tokens AWSPARAMSTR#/appxyz/service1-password
source app.env
./startapp
```

by default the output path is `app.env` relative to the exec binary.

this can be overridden by passing in the `--path` param

```bash
genvars --token AWSSECRETS#/appxyz/service1-password --token AWSPARAMSTR#/appxyz/service12-settings --path /some/path/app.env
source /some/path/app.env
./startapp
```

The token is made up of 3 parts:
    - `AWSSECRETS` the strategy identifier to choose at runtime
    - `#` separator 
    - `/path/to/parameter` the actual path to the secret or parameter in the target system e.g. AWS SecretsManager or ParameterStore (it does assume a path like pattern might throw a runtime error if not found)

If contents of the `AWSSECRETS#/appxyz/service1-password` are a string then `service1-password` will be the key and converted to UPPERCASEe.g. 

## Help

More implementations should be easily added with a specific implementation under the strategy interface

TODO: lots more documentation and UNITTESTS


## Download 

<!-- `uname -s`-`uname -m` -->
```bash
curl -L https://github.com/dnitsch/genvars/releases/download/v0.1.0/genvars -o genvars
chmod +x genvars
sudo mv genvars /usr/local/bin
```