package gen

import (
	"errors"
	"fmt"
	"gufeijun/hustgen/config"
	"gufeijun/hustgen/gen/cgen"
	"gufeijun/hustgen/gen/gogen"
	"gufeijun/hustgen/gen/nodegen"
	"gufeijun/hustgen/parse"
	"os"
	"path/filepath"
	"strings"
)

type Generator struct {
	infos          *parse.Symbols
	langGenerators map[string]func(*parse.Symbols, *config.ComplileConfig) error
}

func (g *Generator) langHelp(lang string) error {
	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("do not support language: %s\n", lang))
	builder.WriteString("supported langs: ")
	for lang, _ := range g.langGenerators {
		builder.WriteString(lang)
		builder.WriteByte(' ')
	}
	return errors.New(builder.String())
}

func NewGenerator(infos *parse.Symbols) *Generator {
	return &Generator{
		langGenerators: map[string]func(*parse.Symbols, *config.ComplileConfig) error{
			"c":    cgen.Gen,
			"go":   gogen.Gen,
			"node": nodegen.Gen,
		},
		infos: infos,
	}
}

func (g *Generator) Gen(config *config.ComplileConfig) error {
	var err error
	config.OutDir, err = filepath.Abs(config.OutDir)
	if err != nil {
		return err
	}
	f, ok := g.langGenerators[config.TargetLang]
	if !ok {
		return g.langHelp(config.TargetLang)
	}
	if err := os.MkdirAll(config.OutDir, 0777); err != nil {
		return err
	}
	return f(g.infos, config)
}
