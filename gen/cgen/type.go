package cgen

import (
	"fmt"
	"gufeijun/hustgen/service"
	"strings"
)

var IDLtoCType = map[string]string{
	"int8":    "int8_t",
	"int16":   "int16_t",
	"int32":   "int32_t",
	"int64":   "int64_t",
	"uint8":   "uint8_t",
	"uint16":  "uint16_t",
	"uint32":  "uint32_t",
	"uint64":  "uint64_t",
	"float32": "float",
	"float64": "double",
	"string":  "char*",
}

func toClangType(t *service.Type) string {
	switch t.TypeKind {
	case service.TypeKindNormal:
		return IDLtoCType[t.TypeName]
	case service.TypeKindMessage:
		return fmt.Sprintf("struct %s*", t.TypeName)
	default:
	}
	return ""
}

func buildMethod(method *service.Method) string {
	var builder strings.Builder
	builder.WriteString(toClangType(method.RetType))
	fmt.Fprintf(&builder, " %s_%s(", method.Service.Name, method.MethodName)
	for i, t := range method.ReqTypes {
		if i != 0 {
			builder.WriteString(", ")
		}
		builder.WriteString(toClangType(t))
	}
	builder.WriteString(", error_t*);")
	return builder.String()
}
