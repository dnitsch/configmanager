package cmd

import (
	"fmt"
	"os"

	"github.com/dnitsch/configmanager/internal/config"
	"github.com/spf13/cobra"
)

var (
	verbose          bool
	tokenSeparator   string
	keySeparator     string
	configmanagerCmd = &cobra.Command{
		Use:   config.SELF_NAME,
		Short: fmt.Sprintf("%s CLI for retrieving and inserting config or secret variables", config.SELF_NAME),
		Long: fmt.Sprintf(`%s CLI for retrieving config or secret variables.
		Using a specific tokens as an array item`, config.SELF_NAME),
	}
)

func Execute() {
	if err := configmanagerCmd.Execute(); err != nil {
		fmt.Errorf("cli error: %v", err)
		os.Exit(1)
	}
	os.Exit(0)
}

func init() {
	configmanagerCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Verbosity level")
	configmanagerCmd.PersistentFlags().StringVarP(&tokenSeparator, "token-separator", "s", "#", "Separator to use to mark concrete store and the key within it")
	configmanagerCmd.PersistentFlags().StringVarP(&keySeparator, "key-separator", "k", "|", "Separator to use to mark a key look up in a map. e.g. AWSSECRETS#/token/map|key1")
}
