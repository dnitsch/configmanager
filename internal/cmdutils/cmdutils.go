// pacakge Cmdutils
//
// Wraps around the ConfigManager library
// with additional postprocessing capabilities for
// output management when used with cli flags.
package cmdutils

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/dnitsch/configmanager/internal/config"
	"github.com/dnitsch/configmanager/pkg/generator"
	"github.com/dnitsch/configmanager/pkg/log"
)

type configManagerIface interface {
	RetrieveWithInputReplaced(input string) (string, error)
	Retrieve(tokens []string) (generator.ParsedMap, error)
	GeneratorConfig() *config.GenVarsConfig
}

type CmdUtils struct {
	logger        log.ILogger
	configManager configManagerIface
	writer        io.Writer
}

func New(confManager configManagerIface, logger log.ILogger) *CmdUtils {
	return &CmdUtils{
		logger:        logger,
		configManager: confManager,
		writer:        os.Stdout, // default writer
	}
}

func (cmd *CmdUtils) WithWriter(w io.Writer) *CmdUtils {
	cmd.writer = w
	return cmd
}

// GenerateFromTokens is a helper cmd method to call from retrieve command
func (c *CmdUtils) GenerateFromCmd(tokens []string, output string) error {
	err := c.setWriter(output)
	if err != nil {
		return err
	}
	// defer c.writer.Close()
	return c.generateFromToken(tokens)
}

// generateFromToken
func (c *CmdUtils) generateFromToken(tokens []string) error {
	pm, err := c.configManager.Retrieve(tokens)
	if err != nil {
		// return full error to terminal if no tokens were parsed
		if len(pm) < 1 {
			return err
		}
		// else log error only
		c.logger.Error("%e", err)
	}
	// Conver to ExportVars and flush to file
	pp := &PostProcessor{ProcessedMap: pm, Config: c.configManager.GeneratorConfig()}
	pp.ConvertToExportVar()
	return pp.FlushOutToFile(c.writer)
}

// Generate a replaced string from string input command
//
// returns a non empty string if move of temp file is required
func (c *CmdUtils) GenerateStrOut(input, output string) error {

	// outputs and inputs match and are file paths
	if input == output {
		c.logger.Debug("overwrite mode on")

		// create a temp file
		tempfile, err := os.CreateTemp(os.TempDir(), "configmanager")
		if err != nil {
			return err
		}
		defer os.Remove(tempfile.Name())
		c.logger.Debug("tmp file created: %s", tempfile.Name())
		if err := c.setWriter(tempfile.Name()); err != nil {
			return err
		}
		// defer c.writer.Close()
		return c.generateFromStrOutOverwrite(input, tempfile.Name())
	}

	err := c.setWriter(output)
	if err != nil {
		return err
	}

	// defer c.writer.Close()

	return c.generateFromStrOut(input)
}

// generateFromStrOut
func (c *CmdUtils) generateFromStrOut(input string) error {
	f, err := os.Open(input)
	if err != nil {
		if perr, ok := err.(*os.PathError); ok {
			c.logger.Debug("input is not a valid file path: %v, falling back on using the string directly", perr)
			// is actual string parse and write out to location
			return c.generateStrOutFromInput(strings.NewReader(input), c.writer)
		}
		return err
	}
	defer f.Close()

	return c.generateStrOutFromInput(f, c.writer)
}

// generateFromStrOutOverwrite uses the same file for input as output
// requires additional consideration and must create a temp file
// and then write contents from temp to actual target
// otherwise, two open file operations would be targeting same descriptor
// will cause issues and inconsistent writes
func (c *CmdUtils) generateFromStrOutOverwrite(input, outtemp string) error {

	f, err := os.Open(input)
	if err != nil {
		return err
	}
	defer f.Close()

	if err := c.generateStrOutFromInput(f, c.writer); err != nil {
		return err
	}
	tr, err := os.ReadFile(outtemp)
	if err != nil {
		return err
	}
	// move temp file to output path
	return os.WriteFile(c.configManager.GeneratorConfig().OutputPath(), tr, 0644)
}

// generateStrOutFromInput takes a reader and writer as input
func (c *CmdUtils) generateStrOutFromInput(input io.Reader, output io.Writer) error {

	b, err := io.ReadAll(input)
	if err != nil {
		return err
	}
	str, err := c.configManager.RetrieveWithInputReplaced(string(b))
	if err != nil {
		return err
	}
	pp := &PostProcessor{}

	return pp.StrToFile(output, str)
}

func (c *CmdUtils) setWriter(outputpath string) error {
	// empty output path means StdOut
	if outputpath != "stdout" {
		f, err := os.OpenFile(outputpath, os.O_WRONLY|os.O_CREATE, 0644)
		if err != nil {
			return err
		}
		c.writer = f
	}
	return nil
}

// UploadTokensWithVals takes in a map of key/value pairs and uploads them
func (c *CmdUtils) UploadTokensWithVals(tokens map[string]string) error {
	return fmt.Errorf("notYetImplemented")
}
