package cmd

import (
	"fmt"
	"os"

	"github.com/dnitsch/configmanager/internal/config"
	"github.com/spf13/cobra"
)

var (
	Version  string = "0.0.1"
	Revision string = "1111aaaa"
)

func init() {
	configmanagerCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: fmt.Sprintf("Get version number %s", config.SELF_NAME),
	Long:  `Version and Revision number of the installed CLI`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Version: %s\nRevision: %s\n", Version, Revision)
		// util.CleanExit()
		os.Exit(0)
	},
}
