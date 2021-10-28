package utils

import "gufeijun/hustgen/service"

func TraverseMethod(callback func(method *service.Method) (end bool)) {
	for _, s := range service.GlobalAsset.Services {
		for _, method := range s.Methods {
			if end := callback(method); end {
				return
			}
		}
	}
}

func TraverseReqArgs(callback func(t *service.Type) (end bool)) {
	for _, s := range service.GlobalAsset.Services {
		for _, method := range s.Methods {
			for _, t := range method.ReqTypes {
				if end := callback(t); end {
					return
				}
			}
		}
	}
}

func TraverseRespArgs(callback func(t *service.Type) (end bool)) {
	for _, s := range service.GlobalAsset.Services {
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
