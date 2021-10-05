package main

import (
	"flag"
	"fmt"
	"gufeijun/hustgen/config"
	"gufeijun/hustgen/gen"
	"gufeijun/hustgen/parse"
)

func main() {
	config := parseConfig()
	args := flag.Args()
	var err error
	for _, srcIDL := range args {
		if err = parse.NewParser(srcIDL).Parse(); err != nil {
			break
		}
		config.SrcIDL = srcIDL
		if err = gen.NewGenerator().Gen(config); err != nil {
			break
		}
	}
	if err != nil {
		fmt.Println(err)
	}
	return
}

func parseConfig() *config.ComplileConfig {
	lang := flag.String("lang", "go", "the target languege the IDL will be compliled to")
	dir := flag.String("dir", "gfj", "the dirpath where the generated source code files will be placed")
	flag.Parse()
	return &config.ComplileConfig{
		TargetLang: *lang,
		OutDir:     *dir,
	}
}
