package cmd

import (
	"fmt"

	"github.com/dnitsch/configmanager"
	"github.com/dnitsch/configmanager/internal/cmdutils"
	"github.com/spf13/cobra"
)

var (
	input                string
	retrieveFromStrInput = &cobra.Command{
		Use:     "string-input",
		Aliases: []string{"fromstr", "getfromstr"},
		Short:   `Retrieves all found token values in a specified string input`,
		Long:    `Retrieves all found token values in a specified string input and optionally writes to a file or to stdout in a bash compliant`,
		RunE:    retrieveFromStr,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			// if len(input) < 1 && !getStdIn() {
			if len(input) < 1 {
				return fmt.Errorf("must include input")
			}
			return nil
		},
	}
)

func init() {
	retrieveFromStrInput.PersistentFlags().StringVarP(&input, "input", "i", "", `Path to file which contents will be read in or the contents of a string 
inside a variable to be searched for tokens. 
If value is a valid path it will open it if not it will accept the string as an input. 
e.g. -i /some/file or -i $"(cat /som/file)", are both valid input values`)
	retrieveFromStrInput.MarkPersistentFlagRequired("input")
	retrieveFromStrInput.PersistentFlags().StringVarP(&path, "path", "p", "./app.env", `Path where to write out the 
replaced a config/secret variables. Special value of stdout can be used to return the output to stdout e.g. -p stdout, 
unix style output only`)
	// 	retrieveFromStrInput.PersistentFlags().BoolVarP(&overwriteinputfile, "overwrite", "o", false, `Writes the outputs of the templated file
	// to a the same location as the input file path`)
	configmanagerCmd.AddCommand(retrieveFromStrInput)
}

func retrieveFromStr(cmd *cobra.Command, args []string) error {
	cm := configmanager.New()
	cm.Config.WithTokenSeparator(tokenSeparator).WithOutputPath(path).WithKeySeparator(keySeparator)
	return cmdutils.New(cm).GenerateStrOut(input, path)
}
