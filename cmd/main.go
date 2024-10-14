package main

import (
	"context"
	"os"

	cfgmgr "github.com/dnitsch/configmanager/cmd/configmanager"
	"github.com/dnitsch/configmanager/pkg/log"
)

func main() {
	// leveler := &slog.LevelVar{}
	// logger := log.slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: leveler}))
	logger := log.New(os.Stderr)
	cmd := cfgmgr.NewRootCmd(logger)
	if err := cmd.Execute(context.Background()); err != nil {
		logger.Error("cli error: %v", err)
	}
}
