package cmd

import (
	"fmt"
	"os"

	"github.com/dnitsch/configmanager/internal/config"
	"github.com/dnitsch/configmanager/internal/utils"
	"github.com/dnitsch/configmanager/pkg/generator"
	"github.com/spf13/cobra"
)

// Default empty string array
var tokenArray []string

var (
	tokens           []string
	path             string
	tokenSeparator   string
	configmanagerCmd = &cobra.Command{
		Short: fmt.Sprintf("%s CLI for retrieving config or secret variables", config.SELF_NAME),
		Long: fmt.Sprintf(`%s CLI for retrieving config or secret variables.
		Using a specific tokens as an array item`, config.SELF_NAME),
		RunE: cfgRun,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if len(tokens) < 1 {
				return fmt.Errorf("must include at least 1 token")
			}
			return nil
		},
	}
)

func Execute() {
	if err := configmanagerCmd.Execute(); err != nil {
		fmt.Errorf("cli error: %e", err)
		os.Exit(1)
	}
	os.Exit(0)
}

func init() {
	configmanagerCmd.PersistentFlags().StringArrayVarP(&tokens, "token", "t", tokenArray, "Token pointing to a config/secret variable. This can be specified multiple times.")
	configmanagerCmd.PersistentFlags().StringVarP(&path, "path", "p", "./app.env", "Path where to write out the replaced a config/secret variables")
	configmanagerCmd.PersistentFlags().StringVarP(&tokenSeparator, "token-separator", "s", "#", "Separator to use to mark concrete store and the key within it")
}

func cfgRun(cmd *cobra.Command, args []string) error {
	err := utils.GenerateTokens(generator.GenVarsConfig{Outpath: path, TokenSeparator: tokenSeparator}, tokens)
	if err != nil {
		return err
	}
	return nil
}
