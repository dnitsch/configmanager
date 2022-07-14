package generator

import (
	"fmt"

	"github.com/dnitsch/configmanager/pkg/log"
)

// GenerateFromTokens is a helper cmd method to call from retrieveCmd
func (gv *GenVars) GenerateFromCmd(tokens []string) error {
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
	log.Infof("Vars written to: %s", f)
	return nil
}

//UploadTokensWithVals takes in a map of key/value pairs and uploads them
func (gv *GenVars) UploadTokensWithVals(tokens map[string]string) error {
	return fmt.Errorf("notYetImplemented")
}
