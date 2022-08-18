// command line utils
package utils

import (
	"fmt"
	"os"

	"github.com/dnitsch/configmanager"
	"github.com/dnitsch/configmanager/pkg/generator"
	"github.com/dnitsch/configmanager/pkg/log"
)

// GenerateFromTokens is a helper cmd method to call from retrieve command
func GenerateFromCmd(gv *generator.GenVars, tokens []string) error {
	_, err := gv.Generate(tokens)
	if err != nil {
		log.Errorf("%e", err)
		return err
	}
	// Conver to ExportVars
	gv.ConvertToExportVar()

	w, err := writer(gv.ConfigOutputPath())
	if err != nil {
		return err
	}
	defer w.Close()

	return gv.FlushToFile(w)
}

// Generate a replaced string from string input command
func GenerateStrOut(gv *generator.GenVars, input string) error {
	c := configmanager.ConfigManager{}

	str, err := c.RetrieveWithInputReplaced(input, *gv.Config())
	if err != nil {
		return err
	}

	w, err := writer(gv.ConfigOutputPath())
	if err != nil {
		return err
	}
	defer w.Close()
	return gv.StrToFile(w, str)
}

func writer(outputpath string) (*os.File, error) {
	if outputpath == "stdout" {
		return os.Stdout, nil
	} else {
		return os.OpenFile(outputpath, os.O_WRONLY|os.O_CREATE, 0644)
	}
}

//UploadTokensWithVals takes in a map of key/value pairs and uploads them
func UploadTokensWithVals(tokens map[string]string) error {
	return fmt.Errorf("notYetImplemented")
}
