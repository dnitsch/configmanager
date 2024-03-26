package cmd

import (
	"fmt"

	"github.com/dnitsch/configmanager"
	"github.com/dnitsch/configmanager/internal/cmdutils"
	"github.com/spf13/cobra"
)

// Default empty string array
var tokenArray []string

var (
	tokens      []string
	path        string
	retrieveCmd = &cobra.Command{
		Use:     "retrieve",
		Aliases: []string{"r", "fetch", "get"},
		Short:   `Retrieves a value for token(s) specified`,
		Long:    `Retrieves a value for token(s) specified and optionally writes to a file or to stdout in a bash compliant export KEY=VAL syntax`,
		RunE:    retrieveRun,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if len(tokens) < 1 {
				return fmt.Errorf("must include at least 1 token")
			}
			return nil
		},
	}
)

func init() {
	retrieveCmd.PersistentFlags().StringArrayVarP(&tokens, "token", "t", tokenArray, "Token pointing to a config/secret variable. This can be specified multiple times.")
	retrieveCmd.MarkPersistentFlagRequired("token")
	retrieveCmd.PersistentFlags().StringVarP(&path, "path", "p", "./app.env", "Path where to write out the replaced a config/secret variables. Special value of stdout can be used to return the output to stdout e.g. -p stdout, unix style output only")
	configmanagerCmd.AddCommand(retrieveCmd)
}

func retrieveRun(cmd *cobra.Command, args []string) error {
	cm := configmanager.New()
	cm.Config.WithTokenSeparator(tokenSeparator).WithOutputPath(path).WithKeySeparator(keySeparator)
	return cmdutils.New(cm).GenerateFromCmd(tokens, path)
}
