package cmd

import (
	"fmt"

	"github.com/dnitsch/configmanager/cmd/utils"
	"github.com/dnitsch/configmanager/pkg/generator"
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
	retrieveFromStrInput.PersistentFlags().StringVarP(&input, "input", "i", "", "Contents of a string inside a variable to be searched for tokens. e.g. -i $(cat /som/file)")
	retrieveFromStrInput.MarkPersistentFlagRequired("input")
	retrieveFromStrInput.PersistentFlags().StringVarP(&path, "path", "p", "./app.env", "Path where to write out the replaced a config/secret variables. Special value of stdout can be used to return the output to stdout e.g. -p stdout, unix style output only")
	configmanagerCmd.AddCommand(retrieveFromStrInput)
}

func retrieveFromStr(cmd *cobra.Command, args []string) error {
	conf := generator.NewConfig().WithTokenSeparator(tokenSeparator).WithOutputPath(path)
	gv := generator.NewGenerator().WithConfig(conf)
	return utils.GenerateStrOut(gv, input)
}
