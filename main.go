package main

import (
	"flag"
	"fmt"
	"gufeijun/hustgen/parse"
)

var (
	printVersion = flag.Bool("version", false, "print program build version")
	lang         = flag.String("lang", "c", "the target languege the IDL will be compliled to")
	dir          = flag.String("dir", "gfj", "the dirpath where the generated source code files will be placed")
)

func init() {
	flag.Parse()
}

// func testToken() {
// 	data, err := ioutil.ReadFile("./test.gfj")
// 	if err != nil {
// 		panic(err)
// 	}
// 	l, err := parse.NewLexer(data)
// 	if err != nil {
// 		panic(err)
// 	}
// 	for {
// 		l.GetNextToken()
// 		token := l.GetToken()
// 		if token.Kind == parse.T_EOF {
// 			break
// 		}
// 		switch token.Kind {
// 		case parse.T_ID:
// 			fmt.Printf("<ID,%s> %dth line", token.Value, token.Line)
// 		case parse.T_CRLF:
// 			fmt.Printf("<CRLF,-> %dth line", token.Line)
// 		case parse.T_COMMA:
// 			fmt.Printf("<COMMA,-> %dth line", token.Line)
// 		case parse.T_MESSAGE:
// 			fmt.Printf("<message,-> %dth line", token.Line)
// 		case parse.T_SERVICE:
// 			fmt.Printf("<service,-> %dth line", token.Line)
// 		case parse.T_LEFTBRACE:
// 			fmt.Printf("<{,-> %dth line", token.Line)
// 		case parse.T_SEMICOLON:
// 			fmt.Printf("<;,-> %dth line", token.Line)
// 		case parse.T_RIGHTBRACE:
// 			fmt.Printf("<},-> %dth line", token.Line)
// 		case parse.T_LEFTBRACKET:
// 			fmt.Printf("<(,-> %dth line", token.Line)
// 		case parse.T_RIGHTBRACKET:
// 			fmt.Printf("<),-> %dth line", token.Line)
// 		}
// 		fmt.Println()
// 	}

// }

// func testPreHandle() {
// 	data, err := ioutil.ReadFile("./test.gfj")
// 	if err != nil {
// 		panic(err)
// 	}
// 	fmt.Printf("-----------------------------------\n")
// 	fmt.Printf("%s\n", data)
// 	l, err := parse.NewLexer(data)
// 	if err != nil {
// 		panic(err)
// 	}
// 	fmt.Printf("-----------------------------------\n")
// 	fmt.Printf("%s\n", l.GetCode())
// }

func testInfos() {
	parser := parse.NewParser("./test.gfj")
	if err := parser.Parse(); err != nil {
		panic(err)
	}
	infos := parser.Infos
	msgs := infos.Messages
	srvs := infos.Services
	for _, msg := range msgs {
		fmt.Printf("message %s{\n", msg.Name)
		for _, mem := range msg.Mems {
			fmt.Printf("\t%s %s\n", mem.MemType.Name, mem.MemName)
		}
		fmt.Println("}")
	}
	for _, srv := range srvs {
		fmt.Printf("service %s{\n", srv.Name)
		for _, met := range srv.Methods {
			fmt.Printf("\t%s %s(", met.RetType.Name, met.Name)
			for i, arg := range met.ReqTypes {
				fmt.Printf("%s", arg.Name)
				if i != len(met.ReqTypes)-1 {
					fmt.Printf(",")
				}
			}
			fmt.Println(")")
		}
		fmt.Println("}")
	}
}

func main() {
	testInfos()
	// testPreHandle()
	// testToken()

	// conf := &config.ComplileConfig{
	// 	TargetLang:   *lang,
	// 	OutDir:       *dir,
	// 	PrintVersion: *printVersion,
	// }
	// if conf.PrintVersion {
	// 	fmt.Printf("Version: %s\n", config.Version)
	// 	return
	// }
	// args := flag.Args()
	// var err error
	// for _, srcIDL := range args {
	// 	if err = parse.NewParser(srcIDL).Parse(); err != nil {
	// 		break
	// 	}
	// 	conf.SrcIDL = srcIDL
	// 	if err = gen.NewGenerator().Gen(conf); err != nil {
	// 		break
	// 	}
	// }
	// if err != nil {
	// 	fmt.Println(err)
	// }
	// return
}
