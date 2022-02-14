package configmanager

import (
	"fmt"

	"github.com/dnitsch/configmanager/pkg/generator"
)

// Retrieve gets a rawMap from a set implementaion
// will be empty if no matches found
func Retrieve(tokens []string, config *generator.GenVarsConfig) (generator.ParsedMap, error) {
	gv := generator.New()
	gv.WithConfig(config)
	return gv.Generate(tokens)
}

func Insert() error {
	return fmt.Errorf("%s", "NotYetImplemented")
}
