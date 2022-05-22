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
		Use:   "version",
		Short: fmt.Sprintf("Get version number %s", config.SELF_NAME),
		Long:  `Version and Revision number of the installed CLI`,
		Run:   cfgRun,
	}
)

func Execute() {
	if err := configmanagerCmd.Execute(); err != nil {
		fmt.Errorf("Command Errord: %v", err)
		os.Exit(1)
	}
	os.Exit(0)
}

func init() {
	configmanagerCmd.PersistentFlags().StringArrayVarP(&tokens, "token", "t", tokenArray, "Token pointing to a config/secret variable")
	configmanagerCmd.MarkPersistentFlagRequired("token")
	configmanagerCmd.PersistentFlags().StringVarP(&path, "path", "p", "./app.env", "Path where to write out the replaced a config/secret variables")
	configmanagerCmd.PersistentFlags().StringVarP(&tokenSeparator, "token-separator", "s", "#", "Separator to use to mark concrete store and the key within it")
}

func cfgRun(cmd *cobra.Command, args []string) {
	err := utils.GenerateTokens(generator.GenVarsConfig{Outpath: path, TokenSeparator: tokenSeparator}, tokens)
	if err != nil {

		os.Exit(1)
	}
}
