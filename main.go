package main

import (
	"flag"
	"fmt"
	"gufeijun/hustgen/config"
	"gufeijun/hustgen/gen"
	"gufeijun/hustgen/parse"
	"os"
)

var (
	printVersion = flag.Bool("version", false, "print program build version")
	lang         = flag.String("lang", "c", "the target languege the IDL will be compliled to. c, go or node.")
	dir          = flag.String("dir", "gfj", "the dirpath where the generated source code files will be placed")
)

func init() {
	flag.Parse()
}

func main() {
	if len(os.Args) <= 1 {
		fmt.Printf("Usage: %s [options] <file,[file...]>\n", os.Args[0])
		fmt.Printf("Execute \"%s --help\" for more details\n", os.Args[0])
		os.Exit(0)
	}
	conf := &config.ComplileConfig{
		TargetLang:   *lang,
		OutDir:       *dir,
		PrintVersion: *printVersion,
	}
	if conf.PrintVersion {
		fmt.Printf("Version: %s\n", config.Version)
		return
	}
	args := flag.Args()
	var err error
	for _, srcIDL := range args {
		parser := parse.NewParser(srcIDL)
		if err = parser.Parse(); err != nil {
			break
		}
		conf.SrcIDL = srcIDL
		if err = gen.NewGenerator(parser.Infos).Gen(conf); err != nil {
			break
		}
	}
	if err != nil {
		fmt.Println(err)
	}
	return
}
