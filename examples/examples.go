package main

import (
	"fmt"

	"github.com/dnitsch/configmanager"
	"github.com/dnitsch/configmanager/pkg/generator"
)

const DO_STUFF_WITH_VALS_HERE = "connstring:user@%v:host=%s/someschema..."

func main() {
	retrieveExample()
	retrieveStringOut()
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
