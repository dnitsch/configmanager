package configmanager

import (
	"fmt"

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

func Insert() error {
	return fmt.Errorf("%s", "NotYetImplemented")
}
