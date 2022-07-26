// command line utils
package utils

import (
	"fmt"

	"github.com/dnitsch/configmanager"
	"github.com/dnitsch/configmanager/pkg/generator"
	"github.com/dnitsch/configmanager/pkg/log"
)

// GenerateFromTokens is a helper cmd method to call from retrieveCmd
func GenerateFromCmd(gv *generator.GenVars, tokens []string) error {
	_, err := gv.Generate(tokens)
	if err != nil {
		log.Errorf("%e", err)
		return err
	}
	// Conver to ExportVars
	gv.ConvertToExportVar()
	return gv.FlushToFile()
}

func GenerateStrOut(gv *generator.GenVars, input string) error {
	c := configmanager.ConfigManager{}

	str, err := c.RetrieveWithInputReplaced(input, *gv.Config())
	if err != nil {
		return err
	}

	return gv.StrToFile(str)
}

//UploadTokensWithVals takes in a map of key/value pairs and uploads them
func UploadTokensWithVals(tokens map[string]string) error {
	return fmt.Errorf("notYetImplemented")
}
