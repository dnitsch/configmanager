package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	insertKv        map[string]string
	defaultInsertKv = map[string]string{}
	insertCmd       = &cobra.Command{
		Use:     "insert",
		Aliases: []string{"i", "send", "put"},
		Short:   `Retrieves a value for token(s) specified and optionally writes to a file`,
		Long:    `Retrieves a value for token(s) specified and optionally writes to a file`,
		RunE:    insertRun,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if len(tokens) < 1 {
				return fmt.Errorf("must include at least 1 token")
			}
			return nil
		},
	}
)

func init() {
	insertCmd.PersistentFlags().StringToStringVarP(&insertKv, "item", "t", defaultInsertKv, "Token pointing to a config/secret variable. This can be specified multiple times.")
	insertCmd.MarkPersistentFlagRequired("item")
	insertCmd.PersistentFlags().StringVarP(&tokenSeparator, "token-separator", "s", "#", "Separator to use to mark concrete store and the key within it")
	configmanagerCmd.AddCommand(insertCmd)
}

func insertRun(cmd *cobra.Command, args []string) error {

	// conf := generator.NewConfig().WithTokenSeparator(tokenSeparator)
	// err := utils.GenerateTokens(*conf, insertKv)
	// if err != nil {
	// 	return err
	// }
	return nil
}
