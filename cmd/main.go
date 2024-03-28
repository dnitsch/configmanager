package main

import (
	"context"

	cfgmgr "github.com/dnitsch/configmanager/cmd/configmanager"
)

func main() {
	// init loggerHere or in init function
	cfgmgr.Execute(context.Background())
}
