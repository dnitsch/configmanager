package utils

import (
	"github.com/dnitsch/configmanager/pkg/generator"
	"github.com/dnitsch/configmanager/pkg/log"
)

// GenerateTokens is a helper cmd method to call from rootCmd
func GenerateTokens(config generator.GenVarsConfig, tokens []string) error {
	gv := generator.New()
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
