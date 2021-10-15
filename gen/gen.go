package gen

import (
	"errors"
	"fmt"
	"gufeijun/hustgen/config"
	"gufeijun/hustgen/gen/gogen"
	"os"
	"path/filepath"
	"strings"
)

type Generator struct {
	langs map[string]func(*config.ComplileConfig) error
}

func (g *Generator) langHelp(lang string) error {
	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("do not support language: %s\n", lang))
	builder.WriteString("supported langs: ")
	for key, _ := range g.langs {
		builder.WriteString(key)
		builder.WriteByte(' ')
	}
	return errors.New(builder.String())
}

func NewGenerator() *Generator {
	return &Generator{
		langs: map[string]func(*config.ComplileConfig) error{
			"go": gogen.Gen,
		},
	}
}

func (g *Generator) Gen(config *config.ComplileConfig) error {
	var err error
	config.OutDir, err = filepath.Abs(config.OutDir)
	if err != nil {
		return err
	}
	f, ok := g.langs[config.TargetLang]
	if !ok {
		return g.langHelp(config.TargetLang)
	}
	if err := os.MkdirAll(config.OutDir, 0777); err != nil {
		return err
	}
	return f(config)
}
