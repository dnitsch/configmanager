package main

import (
	"context"
	"flag"
	"fmt"

	"github.com/dnitsch/genvars"
)

type tokenArray []string

func (i *tokenArray) String() string {
	return ""
}

func (i *tokenArray) Set(value string) error {
	*i = append(*i, value)
	return nil
}

var (
	tokens tokenArray
	path   string
)

func main() {
	flag.Parse()
	gv := genvars.NewGenVars(path, context.TODO())
	gv.WithConfig(&genvars.GenVarsConfig{Outpath: path})
	path, err := gv.Generate(tokens)
	if err != nil {
		fmt.Errorf("%e", err)
	}
	fmt.Printf(path)

}

func init() {
	flag.Var(&tokens, "tokens", "token value to look for in specifc implementation")
	flag.StringVar(&path, "path", "./app.env", "Path to write the sourceable file to")
}
