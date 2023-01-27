// command line utils
// testable methods that wrap around the low level
// implementation when invoked from the cli method.
package cmdutils

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/dnitsch/configmanager/pkg/generator"
	"github.com/dnitsch/configmanager/pkg/log"
)

type confMgrRetrieveWithInputReplacediface interface {
	RetrieveWithInputReplaced(input string, config generator.GenVarsConfig) (string, error)
}

type CmdUtils struct {
	cfgmgr    confMgrRetrieveWithInputReplacediface
	generator generator.GenVarsiface
}

func New(gv generator.GenVarsiface, confManager confMgrRetrieveWithInputReplacediface) *CmdUtils {
	return &CmdUtils{
		cfgmgr:    confManager,
		generator: gv,
	}
}

// GenerateFromTokens is a helper cmd method to call from retrieve command
func (c *CmdUtils) GenerateFromCmd(tokens []string, output string) error {
	w, err := writer(output)
	if err != nil {
		return err
	}
	defer w.Close()
	return c.generateFromToken(tokens, w)
}

// generateFromToken
func (c *CmdUtils) generateFromToken(tokens []string, w io.Writer) error {
	pm, err := c.generator.Generate(tokens)
	if err != nil {
		// return full error to terminal if no tokens were parsed
		if len(pm) < 1 {
			return err
		}
		// else log error only
		log.Errorf("%e", err)
	}
	// Conver to ExportVars and flush to file
	return c.generator.FlushToFile(w, c.generator.ConvertToExportVar())
}

// Generate a replaced string from string input command
//
// returns a non empty string if move of temp file is required
func (c *CmdUtils) GenerateStrOut(input, output string) error {

	// outputs and inputs match and are file paths
	if input == output {
		log.Debugf("overwrite mode on")

		// create a temp file
		tempfile, err := os.CreateTemp(os.TempDir(), "configmanager")
		if err != nil {
			return err
		}
		defer os.Remove(tempfile.Name())
		log.Debugf("tmp file created: %s", tempfile.Name())
		outtmp, err := writer(tempfile.Name())
		if err != nil {
			return err
		}
		defer outtmp.Close()
		return c.generateFromStrOutOverwrite(input, tempfile.Name(), outtmp)
	}

	out, err := writer(output)
	if err != nil {
		return err
	}

	defer out.Close()

	return c.generateFromStrOut(input, out)
}

// generateFromStrOut
func (c *CmdUtils) generateFromStrOut(input string, output io.Writer) error {
	f, err := os.Open(input)
	if err != nil {
		if perr, ok := err.(*os.PathError); ok {
			log.Debugf("input is not a valid file path: %v, falling back on using the string directly", perr)
			// is actual string parse and write out to location
			return c.generateStrOutFromInput(strings.NewReader(input), output)
		}
		return err
	}
	defer f.Close()

	return c.generateStrOutFromInput(f, output)
}

// generateFromStrOutOverwrite uses the same file for input as output
// requires additional consideration and must create a temp file
// and then write contents from temp to actual target
// otherwise, two open file operations would be targeting same descriptor
// will cause issues and inconsistent writes
func (c *CmdUtils) generateFromStrOutOverwrite(input, outtemp string, outtmp io.Writer) error {

	f, err := os.Open(input)
	if err != nil {
		return err
	}
	defer f.Close()

	if err := c.generateStrOutFromInput(f, outtmp); err != nil {
		return err
	}
	tr, err := os.ReadFile(outtemp)
	if err != nil {
		return err
	}

	// move temp file to output path
	return os.WriteFile(c.generator.Config().OutputPath(), tr, 0644)
}

// generateStrOutFromInput takes a reader and writer as input
func (c *CmdUtils) generateStrOutFromInput(input io.Reader, output io.Writer) error {

	b, err := io.ReadAll(input)
	if err != nil {
		return err
	}
	str, err := c.cfgmgr.RetrieveWithInputReplaced(string(b), *c.generator.Config())
	if err != nil {
		return err
	}

	return c.generator.StrToFile(output, str)
}

func writer(outputpath string) (*os.File, error) {
	if outputpath == "stdout" {
		return os.Stdout, nil
	}
	return os.OpenFile(outputpath, os.O_WRONLY|os.O_CREATE, 0644)
}

// UploadTokensWithVals takes in a map of key/value pairs and uploads them
func (c *CmdUtils) UploadTokensWithVals(tokens map[string]string) error {
	return fmt.Errorf("notYetImplemented")
}
