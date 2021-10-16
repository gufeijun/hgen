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
