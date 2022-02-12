package configmanager

import (
	"context"

	"github.com/dnitsch/configmanager/pkg/generator"
	"github.com/dnitsch/configmanager/pkg/log"
)

func Retrieve(tokens []string, ctx context.Context) (generator.ParsedMap, error) {
	gv := generator.NewGenVars(context.TODO())
	gv.WithConfig(&generator.GenVarsConfig{})
	rawmap, err := gv.Generate(tokens)
	if err != nil {
		log.Errorf("%e", err)
		return err
	}
	log.Infof("%+v", rawmap)
}
