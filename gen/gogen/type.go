package gogen

import (
	"fmt"
	"gufeijun/hustgen/service"
	"strings"
)

var toGlangMap1 = map[string]string{
	"istream": "io.Reader",
	"ostream": "io.Writer",
	"stream":  "io.ReadWriter",
}

var toGlangMap2 = map[string]string{
	"istream": "io.ReadCloser",
	"ostream": "io.WriteCloser",
	"stream":  "io.ReadWriteCloser",
}

func buildCallArgs(reqTypes []*service.Type) (callArgs []*CallArg) {
	for i, t := range reqTypes {
		callArgs = append(callArgs, &CallArg{
			TypeKind: t.TypeKind,
			TypeName: t.TypeName,
			Data:     fmt.Sprintf("arg%d", i+1),
		})
	}
	return
}

func buildReturn(retType *service.Type) string {
	if retType.TypeName == "void" {
		return "return err"
	}
	if retType.TypeKind == service.TypeKindMessage {
		return fmt.Sprintf(`res = new(%s)
	return res, json.Unmarshal(resp.([]byte), res)
`, retType.TypeName)
	}
	return fmt.Sprintf("return resp.(%s),err", toGolangType(retType, true))
}

func buildResponseArg(retType *service.Type) string {
	if retType.TypeName == "void" {
		return "err error"
	}
	return fmt.Sprintf("res %s, err error", toGolangType(retType, true))
}

func buildRequestArgs(reqTypes []*service.Type) string {
	var builder strings.Builder
	types := toGolangTypes(reqTypes, false)
	for i, t := range types {
		if i != 0 {
			builder.WriteString(", ")
		}
		fmt.Fprintf(&builder, "arg%d %s", i+1, t)
	}
	return builder.String()
}

func toGolangMethod(m *service.Method) (method string) {
	var builder strings.Builder
	types := toGolangTypes(append([]*service.Type{m.RetType}, m.ReqTypes...), false)
	fmt.Fprintf(&builder, "%s(", m.MethodName)
	if len(types) != 1 {
		for i := 1; i < len(types); i++ {
			if i != 1 {
				builder.WriteString(", ")
			}
			builder.WriteString(types[i])
		}
	}
	builder.WriteString(") ")
	if types[0] == "void" {
		builder.WriteString("error")
	} else if m.RetType.TypeKind == service.TypeKindStream {
		fmt.Fprintf(&builder, "(stream %s, onFinish func(), err error)", types[0])
	} else {
		fmt.Fprintf(&builder, "(%s, error)", types[0])
	}
	return builder.String()
}

func toGolangTypes(ts []*service.Type, closer bool) []string {
	var types []string
	for _, t := range ts {
		types = append(types, toGolangType(t, closer))
	}
	return types
}

func toGolangType(t *service.Type, closer bool) string {
	switch t.TypeKind {
	case service.TypeKindNormal:
		return t.TypeName
	case service.TypeKindMessage:
		return "*" + t.TypeName
	case service.TypeKindStream:
		if closer {
			return toGlangMap2[t.TypeName]
		}
		return toGlangMap1[t.TypeName]
	default:
		return ""
	}
}
