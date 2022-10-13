package configmanager

import (
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/dnitsch/configmanager/pkg/generator"
)

const (
	TERMINATING_CHAR string = `[^\'\"\s\n]`
)

type ConfigManageriface interface {
	Retrieve(tokens []string, config generator.GenVarsConfig) (generator.ParsedMap, error)
	RetrieveWithInputReplaced(input string, config generator.GenVarsConfig) (string, error)
	Insert(force bool) error
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
// using regex to grab a list of found tokens in the given string and return the replaced string
func (c *ConfigManager) RetrieveWithInputReplaced(input string, config generator.GenVarsConfig) (string, error) {
	gv := generator.NewGenerator().WithConfig(&config)
	return retrieveWithInputReplaced(input, gv)
}

func retrieveWithInputReplaced(input string, gv generator.Generatoriface) (string, error) {
	tokens := []string{}
	for k := range generator.VarPrefix {
		matches := regexp.MustCompile(`(?s)`+regexp.QuoteMeta(k)+`.(`+TERMINATING_CHAR+`+)`).FindAllString(input, -1)
		tokens = append(tokens, matches...)
	}

	m, err := retrieve(tokens, gv)

	if err != nil {
		return "", err
	}

	return replaceString(m, input), nil
}

// replaceString fills tokens in a provided input with their actual secret/config values
func replaceString(inputMap generator.ParsedMap, inputString string) string {
	mkeys := make([]string, 0, len(inputMap))
	for k := range inputMap {
		mkeys = append(mkeys, k)
	}

	// order map by keys length
	sort.Slice(mkeys, func(i, j int) bool {
		l1, l2 := len(mkeys[i]), len(mkeys[j])
		if l1 != l2 {
			return l1 > l2
		}
		return mkeys[i] > mkeys[j]
	})

	// ordered values by index
	for _, oval := range mkeys {
		inputString = strings.ReplaceAll(inputString, oval, fmt.Sprint(inputMap[oval]))
	}
	return inputString
}

// KubeControllerSpecHelper is a helper method. It has to be static method
// as an interface cannot have methods with type paramaters yet...
//
// It accepts a DI of configmanager and the config (for testability) to replace all occurences of replaceable tokens inside a Marshalled string of that type
func KubeControllerSpecHelper[T any](inputType T, cm ConfigManageriface, config generator.GenVarsConfig) (*T, error) {
	outType := new(T)
	rawBytes, err := json.Marshal(inputType)
	if err != nil {
		return nil, err
	}

	replaced, err := cm.RetrieveWithInputReplaced(string(rawBytes), config)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal([]byte(replaced), outType); err != nil {
		return nil, err
	}
	return outType, nil
}

// Insert will update
func (c *ConfigManager) Insert(force bool) error {
	return fmt.Errorf("%s", "NotYetImplemented")
}
