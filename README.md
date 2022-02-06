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
genvars --tokens AWSSECRETS#/appxyz/service1-password --tokens AWSPARAMSTR#/appxyz/service12-settings --path /some/path/app.env
source /some/path/app.env
./startapp
```

If contents of the `AWSSECRETS#/appxyz/service1-password` are a string

## Help

More implementations should be easily added with a specific implementation 

TODO: lots more documentation and UNITTESTS