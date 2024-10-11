package main

import (
	"context"
	"log"
	"os"

	cfgmgr "github.com/dnitsch/configmanager/cmd/configmanager"
)

func main() {
	cmd := cfgmgr.NewRootCmd(os.Stdout, os.Stderr)
	if err := cmd.Execute(context.Background()); err != nil {
		log.Fatalf("cli error: %v", err)
	}
}
