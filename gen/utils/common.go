package utils

import "gufeijun/hustgen/parse"

func TraverseMethod(infos *parse.Symbols, callback func(method *parse.Method) (end bool)) {
	for _, s := range infos.Services {
		for _, method := range s.Methods {
			if end := callback(method); end {
				return
			}
		}
	}
}

func TraverseReqArgs(infos *parse.Symbols, callback func(t *parse.Type) (end bool)) {
	for _, s := range infos.Services {
		for _, method := range s.Methods {
			for _, t := range method.ReqTypes {
				if end := callback(t); end {
					return
				}
			}
		}
	}
}

func TraverseRespArgs(infos *parse.Symbols, callback func(t *parse.Type) (end bool)) {
	for _, s := range infos.Services {
		for _, method := range s.Methods {
			if end := callback(method.RetType); end {
				return
			}

		}
	}
}

var TypeLength = map[string]int{
	"int8":    1,
	"int16":   2,
	"int32":   4,
	"int64":   8,
	"uint8":   1,
	"uint16":  2,
	"uint32":  4,
	"uint64":  8,
	"float32": 4,
	"float64": 8,
}
