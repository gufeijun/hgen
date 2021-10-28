package nodegen

import (
	"fmt"
	"gufeijun/hustgen/service"
	"strings"
)

func buildNodeMethod(method *service.Method) *methodDesc {
	var desc strings.Builder
	var signature strings.Builder
	fmt.Fprintf(&signature, "%s(", method.MethodName)
	for i, t := range method.ReqTypes {
		if t.TypeName == "void" {
			break
		}
		fmt.Fprintf(&signature, "arg%d", i+1)
		if i != len(method.ReqTypes)-1 {
			signature.WriteString(", ")
		}
		fmt.Fprintf(&desc, "\n\t// arg%d: %s", i+1, t.TypeName)
	}
	signature.WriteByte(')')
	if name := method.RetType.TypeName; name != "void" {
		fmt.Fprintf(&desc, "\n\t// retn: %s", name)
	}
	return &methodDesc{
		Desc:      desc.String(),
		Signature: signature.String(),
	}
}

func buildMethodsName(s *service.Service) (string, []string) {
	var builder strings.Builder
	var methods []string
	for i, method := range s.Methods {
		fmt.Fprintf(&builder, `"%s"`, method.MethodName)
		methods = append(methods, method.MethodName)
		if i != len(s.Methods)-1 {
			fmt.Fprint(&builder, ", ")
		}
	}
	return builder.String(), methods
}
