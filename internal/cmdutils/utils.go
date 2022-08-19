// command line utils
package utils

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"

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
// returns a non empty string if move of temp file is required
func GenerateStrOut(gv *generator.GenVars, input, output string) error {
	overwriteinputfile := input == output
	tmpout := ""
	if overwriteinputfile {
		log.Debugf("overwrite mode on, overwrite: %v", overwriteinputfile)

		// create a temp file
		tempfile, err := ioutil.TempFile(os.TempDir(), "configmanager")
		if err != nil {
			return err
		}
		defer os.Remove(tempfile.Name())
		log.Debugf("tmp file created: %s", tempfile.Name())
		tmpout = tempfile.Name()
	}

	f, e := os.Open(input)
	if e != nil {
		if perr, ok := e.(*os.PathError); ok {
			log.Debugf("input is not a valid file path: %v, falling back on using the string directly", perr)
			// is actual string parse and write out to location
			return generateStrOutFromInput(gv, strings.NewReader(input), output)
		}
		return e
	}
	defer f.Close()

	if overwriteinputfile {
		if err := generateStrOutFromInput(gv, f, tmpout); err != nil {
			return err
		}
		tr, e := ioutil.ReadFile(tmpout)
		if e != nil {
			return e
		}
		// move temp file to output path
		return os.WriteFile(output, tr, 0644)
	}
	return generateStrOutFromInput(gv, f, output)
}

// generateStrOutFromInput takes a raw string as input and
func generateStrOutFromInput(gv *generator.GenVars, input io.Reader, output string) error {

	c := configmanager.ConfigManager{}

	b, err := io.ReadAll(input)
	if err != nil {
		return err
	}
	str, err := c.RetrieveWithInputReplaced(string(b), *gv.Config())
	if err != nil {
		return err
	}

	w, err := writer(output)
	if err != nil {
		return err
	}
	defer w.Close()
	return gv.StrToFile(w, str)
}

func writer(outputpath string) (*os.File, error) {
	if outputpath == "stdout" {
		return os.Stdout, nil
	}
	return os.OpenFile(outputpath, os.O_WRONLY|os.O_CREATE, 0644)
}

//UploadTokensWithVals takes in a map of key/value pairs and uploads them
func UploadTokensWithVals(tokens map[string]string) error {
	return fmt.Errorf("notYetImplemented")
}
