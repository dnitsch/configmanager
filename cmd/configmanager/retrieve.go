package cmd

import (
	"fmt"

	"github.com/dnitsch/configmanager"
	"github.com/dnitsch/configmanager/internal/cmdutils"
	"github.com/dnitsch/configmanager/pkg/generator"
	"github.com/dnitsch/configmanager/pkg/log"
	"github.com/spf13/cobra"
)

type retrieveTokenFlags struct {
	tokens []string
	path   string
}

func newRetrieveCmd(rootCmd *Root) {
	f := &retrieveTokenFlags{}

	retrieveCmd := &cobra.Command{
		Use:     "retrieve",
		Aliases: []string{"r", "fetch", "get"},
		Short:   `Retrieves a value for token(s) specified`,
		Long:    `Retrieves a value for token(s) specified and optionally writes to a file or to stdout in a bash compliant export KEY=VAL syntax`,
		RunE: func(cmd *cobra.Command, args []string) error {
			outputWriter, err := cmdutils.GetWriter(f.path)
			if err != nil {
				return err
			}
			defer outputWriter.Close()

			cm := configmanager.New(cmd.Context())
			cm.Config.WithTokenSeparator(rootCmd.rootFlags.tokenSeparator).WithOutputPath(f.path).WithKeySeparator(rootCmd.rootFlags.keySeparator)
			gnrtr := generator.NewGenerator(cmd.Context(), func(gv *generator.GenVars) {
				if rootCmd.rootFlags.verbose {
					rootCmd.logger.SetLevel(log.DebugLvl)
				}
				gv.Logger = rootCmd.logger
			}).WithConfig(cm.Config)
			cm.WithGenerator(gnrtr)
			return cmdutils.New(cm, rootCmd.logger, outputWriter).GenerateFromCmd(f.tokens)
		},
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if len(f.tokens) < 1 {
				return fmt.Errorf("must include at least 1 token")
			}
			return nil
		},
	}
	retrieveCmd.PersistentFlags().StringArrayVarP(&f.tokens, "token", "t", []string{}, "Token pointing to a config/secret variable. This can be specified multiple times.")
	retrieveCmd.MarkPersistentFlagRequired("token")
	retrieveCmd.PersistentFlags().StringVarP(&f.path, "path", "p", "./app.env", "Path where to write out the replaced a config/secret variables. Special value of stdout can be used to return the output to stdout e.g. -p stdout, unix style output only")
	rootCmd.Cmd.AddCommand(retrieveCmd)
}
