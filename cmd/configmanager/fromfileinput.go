package cmd

import (
	"fmt"

	"github.com/dnitsch/configmanager"
	"github.com/dnitsch/configmanager/internal/cmdutils"
	"github.com/dnitsch/configmanager/pkg/generator"
	"github.com/dnitsch/configmanager/pkg/log"
	"github.com/spf13/cobra"
)

type fromStrFlags struct {
	input string
	path  string
}

func newFromStrCmd(rootCmd *Root) {

	f := &fromStrFlags{}

	fromstrCmd := &cobra.Command{
		Use:     "string-input",
		Aliases: []string{"fromstr", "getfromstr"},
		Short:   `Retrieves all found token values in a specified string input`,
		Long:    `Retrieves all found token values in a specified string input and optionally writes to a file or to stdout in a bash compliant`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cm := configmanager.New(cmd.Context())
			cm.Config.WithTokenSeparator(rootCmd.rootFlags.tokenSeparator).WithOutputPath(f.path).WithKeySeparator(rootCmd.rootFlags.keySeparator)
			if f.path == "stdout" {

			}
			gnrtr := generator.NewGenerator(cmd.Context(), func(gv *generator.GenVars) {
				if rootCmd.rootFlags.verbose {
					rootCmd.logger.SetLevel(log.DebugLvl)
				}
				gv.Logger = rootCmd.logger
			}).WithConfig(cm.Config)
			cm.WithGenerator(gnrtr)
			return cmdutils.New(cm, rootCmd.logger).GenerateStrOut(f.input, f.path)
		},
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if len(f.input) < 1 {
				return fmt.Errorf("must include input")
			}
			return nil
		},
	}

	fromstrCmd.PersistentFlags().StringVarP(&f.input, "input", "i", "", `Path to file which contents will be read in or the contents of a string 
	inside a variable to be searched for tokens. 
	If value is a valid path it will open it if not it will accept the string as an input. 
	e.g. -i /some/file or -i $"(cat /som/file)", are both valid input values`)
	fromstrCmd.MarkPersistentFlagRequired("input")
	fromstrCmd.PersistentFlags().StringVarP(&f.path, "path", "p", "./app.env", `Path where to write out the 
	replaced a config/secret variables. Special value of stdout can be used to return the output to stdout e.g. -p stdout, 
	unix style output only`)
	// 	retrieveFromStrInput.PersistentFlags().BoolVarP(&overwriteinputfile, "overwrite", "o", false, `Writes the outputs of the templated file
	// to a the same location as the input file path`)
	rootCmd.Cmd.AddCommand(fromstrCmd)
}
