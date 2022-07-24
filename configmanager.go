package configmanager

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/dnitsch/configmanager/pkg/generator"
)

type ConfigManager interface {
	Retrieve(tokens []string, config generator.GenVarsConfig) (generator.ParsedMap, error)
	Insert()
}

// Retrieve gets a rawMap from a set implementation
// will be empty if no matches found
func Retrieve(tokens []string, config generator.GenVarsConfig) (generator.ParsedMap, error) {
	gv := generator.NewGenerator()
	gv.WithConfig(&config)
	return gv.Generate(tokens)
}

// RetrieveWithInputReplaced parses given input against all possible token strings
// using regex to grab a list of found tokens in the given string
func RetrieveWithInputReplaced(input string, config generator.GenVarsConfig) (string, error) {
	tokens := []string{}
	for k := range generator.VarPrefix {
		matches := regexp.MustCompile(`(?s)`+regexp.QuoteMeta(k)+`.([^\"]+)`).FindAllString(input, -1)
		tokens = append(tokens, matches...)
	}

	cnf := generator.GenVarsConfig{}
	m, err := Retrieve(tokens, cnf)

	if err != nil {
		return "", err
	}

	return replaceString(m, input), nil
}

func replaceString(inputMap generator.ParsedMap, inputString string) string {
	for oldVal, newVal := range inputMap {
		inputString = strings.ReplaceAll(inputString, oldVal, fmt.Sprint(newVal))
	}
	return inputString
}

func Insert() error {
	return fmt.Errorf("%s", "NotYetImplemented")
}
