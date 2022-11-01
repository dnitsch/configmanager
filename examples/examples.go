package main

import (
	"encoding/json"
	"fmt"

	"github.com/dnitsch/configmanager"
	"github.com/dnitsch/configmanager/pkg/generator"
)

const DO_STUFF_WITH_VALS_HERE = "connstring:user@%v:host=%s/someschema..."

func main() {
	retrieveExample()
	retrieveStringOut()
	retrieveYaml()
}

// retrieveExample uses the standard Retrieve method on the API
// this will return generator.ParsedMap which can be later used for more complex use cases
func retrieveExample() {
	cm := &configmanager.ConfigManager{}
	cnf := generator.NewConfig()

	pm, err := cm.Retrieve([]string{"token1", "token2"}, *cnf)

	if err != nil {
		panic(err)
	}

	// put in a loop for many config params
	// or use the helper methods to return a yaml replaced struct
	//
	if pwd, ok := pm["token1"]; ok {
		if host, ok := pm["token2"]; ok {
			fmt.Println(fmt.Sprintf(DO_STUFF_WITH_VALS_HERE, pwd, fmt.Sprintf("%s", host)))
		}
	}
}

// retrieveStringOut accepts a string as an input
func retrieveStringOut() {
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

// ConfigTokenReplace uses configmanager to replace all occurences of
// replaceable tokens inside a []byte
// this is a re-useable method on all controllers
// will just ignore any non specs without tokens
func SpecConfigTokenReplace[T any](inputType T) (*T, error) {
	outType := new(T)
	rawBytes, err := json.Marshal(inputType)
	if err != nil {
		return nil, err
	}

	cm := configmanager.ConfigManager{}

	// use custom token separator
	// inline with
	cnf := generator.NewConfig().WithTokenSeparator("://")

	replaced, err := cm.RetrieveWithInputReplaced(string(rawBytes), *cnf)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal([]byte(replaced), outType); err != nil {
		return nil, err
	}
	return outType, nil
}

// Example using a helper method
func retrieveYaml() {
	type config struct {
		DbHost   string `yaml:"dbhost"`
		Username string `yaml:"user"`
		Password string `yaml:"pass"`
	}
	configMarshalled := `
user: AWSPARAMSTR:///int-test/pocketbase/config|user
pass: AWSPARAMSTR:///int-test/pocketbase/config|pwd
dbhost: AWSPARAMSTR:///int-test/pocketbase/config|host
`

	cm := &configmanager.ConfigManager{}
	// use custom token separator inline with future releases
	cmConf := generator.NewConfig().WithTokenSeparator("://")
	appConf, err := configmanager.RetrieveUnmarshalledFromYaml([]byte(configMarshalled), &config{}, cm, *cmConf)
	if err != nil {
		panic(err)
	}
	fmt.Println(appConf.DbHost)
	fmt.Println(appConf.Username)
	fmt.Println(appConf.Password)
}
