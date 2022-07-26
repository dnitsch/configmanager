package configmanager

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/dnitsch/configmanager/pkg/generator"
)

type ConfigManageriface interface {
	Retrieve(tokens []string, config generator.GenVarsConfig) (generator.ParsedMap, error)
	RetrieveWithInputReplaced(input string, config generator.GenVarsConfig) (string, error)
	Insert()
}

type ConfigManager struct{}

// Retrieve gets a rawMap from a set implementation
// will be empty if no matches found
func (c *ConfigManager) Retrieve(tokens []string, config generator.GenVarsConfig) (generator.ParsedMap, error) {
	gv := generator.NewGenerator().WithConfig(&config)
	return retrieve(tokens, gv)
}

func retrieve(tokens []string, gv generator.Generatoriface) (generator.ParsedMap, error) {
	return gv.Generate(tokens)
}

// RetrieveWithInputReplaced parses given input against all possible token strings
// using regex to grab a list of found tokens in the given string
func (c *ConfigManager) RetrieveWithInputReplaced(input string, config generator.GenVarsConfig) (string, error) {
	gv := generator.NewGenerator().WithConfig(&config)
	return retrieveWithInputReplaced(input, gv)
}

func retrieveWithInputReplaced(input string, gv generator.Generatoriface) (string, error) {
	tokens := []string{}
	for k := range generator.VarPrefix {
		matches := regexp.MustCompile(`(?s)`+regexp.QuoteMeta(k)+`.([^\'\"\s\n]+)`).FindAllString(input, -1)
		tokens = append(tokens, matches...)
	}

	m, err := retrieve(tokens, gv)

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

func (c *ConfigManager) Insert() error {
	return fmt.Errorf("%s", "NotYetImplemented")
}
