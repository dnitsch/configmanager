package main

import (
	"context"
	"os"

	cfgmgr "github.com/dnitsch/configmanager/cmd/configmanager"
	"github.com/dnitsch/configmanager/pkg/log"
)

func main() {
	logger := log.New(os.Stderr)
	cmd := cfgmgr.NewRootCmd(logger)
	if err := cmd.Execute(context.Background()); err != nil {
		os.Exit(1)
	}
}
