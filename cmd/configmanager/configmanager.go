package main

import (
	"flag"
	"os"

	"github.com/dnitsch/configmanager/pkg/generator"
	"github.com/dnitsch/configmanager/pkg/log"
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
	token          tokenArray
	path           string
	tokenSeparator string
)

func main() {
	flag.Parse()
	gv := generator.New()
	gv.WithConfig(&generator.GenVarsConfig{Outpath: path})
	_, err := gv.Generate(token)
	if err != nil {
		log.Errorf("%e", err)
		os.Exit(1)
	}
	gv.ConvertToExportVar()

	f, err := gv.FlushToFile()
	if err != nil {
		log.Errorf("%e", err)
		os.Exit(1)
	}
	log.Infof("Vars written to: %s\n", f)
	os.Exit(0)
}

func init() {
	flag.Var(&token, "token", "token value to look for in specifc implementation")
	flag.StringVar(&path, "path", "./app.env", "Path to write the sourceable file to")
	flag.StringVar(&tokenSeparator, "tokenseparator", generator.TokenSeparator, "Token Separator symbol to use")
}
