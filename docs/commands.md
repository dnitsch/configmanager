# Commands


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

### retrieve

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

### string-input

Replaces all the occurences of tokens inside strings and writes them back out to a file provided. 

This method can be used with entire application property files such as `application.yml` or `application.properties` for springboot apps or netcore app config in which ever format.

The `fromstr` (alias for `string-input`) respects all indentations so can be used on contents of a file of any type

