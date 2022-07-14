package utils

import (
	"fmt"

	"github.com/dnitsch/configmanager/pkg/generator"
	"github.com/dnitsch/configmanager/pkg/log"
)

// GenerateTokens is a helper cmd method to call from retrieveCmd
func GenerateTokens(config generator.GenVarsConfig, tokens []string) error {
	gv := generator.NewGenerator()
	gv.WithConfig(&config)
	_, err := gv.Generate(tokens)
	if err != nil {
		log.Errorf("%e", err)
		return err
	}

	// Conver to ExportVars
	gv.ConvertToExportVar()

	f, err := gv.FlushToFile()
	if err != nil {
		log.Errorf("%e", err)
		return err
	}
	log.Infof("Vars written to: %s\n", f)
	return nil
}

//UploadTokensWithVals takes in a map of key/value pairs and uploads them
func UploadTokensWithVals(config generator.GenVarsConfig, tokens map[string]string) error {
	return fmt.Errorf("notYetImplemented")
}
