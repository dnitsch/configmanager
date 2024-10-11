package cmd

import (
	"context"
	"fmt"
	"io"

	"github.com/dnitsch/configmanager/internal/config"
	"github.com/spf13/cobra"
)

var (
	Version  string = "0.0.1"
	Revision string = "1111aaaa"
)

type rootCmdFlags struct {
	verbose        bool
	tokenSeparator string
	keySeparator   string
}

type Root struct {
	Cmd        *cobra.Command
	ChannelOut io.Writer
	ChannelErr io.Writer
	rootFlags  *rootCmdFlags
}

func NewRootCmd(channelOut, channelErr io.Writer) *Root {
	rc := &Root{
		Cmd: &cobra.Command{
			Use:   config.SELF_NAME,
			Short: fmt.Sprintf("%s CLI for retrieving and inserting config or secret variables", config.SELF_NAME),
			Long: fmt.Sprintf(`%s CLI for retrieving config or secret variables.
			Using a specific tokens as an array item`, config.SELF_NAME),
			Version: fmt.Sprintf("Version: %s\nRevision: %s\n", Version, Revision),
		},
		ChannelOut: channelOut,
		ChannelErr: channelErr,
		rootFlags:  &rootCmdFlags{},
	}

	rc.Cmd.PersistentFlags().BoolVarP(&rc.rootFlags.verbose, "verbose", "v", false, "Verbosity level")
	rc.Cmd.PersistentFlags().StringVarP(&rc.rootFlags.tokenSeparator, "token-separator", "s", "#", "Separator to use to mark concrete store and the key within it")
	rc.Cmd.PersistentFlags().StringVarP(&rc.rootFlags.keySeparator, "key-separator", "k", "|", "Separator to use to mark a key look up in a map. e.g. AWSSECRETS#/token/map|key1")
	addSubCmds(rc)
	return rc
}

// addSubCmds assigns the subcommands to the parent/root command
func addSubCmds(rootCmd *Root) {
	newFromStrCmd(rootCmd)
	newRetrieveCmd(rootCmd)
	newInsertCmd(rootCmd)
}

func (rc *Root) Execute(ctx context.Context) error {
	return rc.Cmd.ExecuteContext(ctx)
}
