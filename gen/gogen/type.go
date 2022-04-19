package gogen

import (
	"fmt"
	"gufeijun/hustgen/parse"
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

func buildCallArgs(reqTypes []*parse.Type) (callArgs []*CallArg) {
	for i, t := range reqTypes {
		callArgs = append(callArgs, &CallArg{
			TypeKind: t.Kind,
			TypeName: t.Name,
			Data:     fmt.Sprintf("arg%d", i+1),
		})
	}
	return
}

func buildReturn(retType *parse.Type) string {
	if retType.Name == "void" {
		return "return err"
	}
	if retType.Kind == parse.TypeKindMessage {
		return fmt.Sprintf(`res = new(%s)
	return res, json.Unmarshal(resp.([]byte), res)
`, retType.Name)
	}
	return fmt.Sprintf("return resp.(%s),err", toGolangType(retType, true))
}

func buildResponseArg(retType *parse.Type) string {
	if retType.Name == "void" {
		return "err error"
	}
	return fmt.Sprintf("res %s, err error", toGolangType(retType, true))
}

func buildRequestArgs(reqTypes []*parse.Type) string {
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

func toGolangMethod(m *parse.Method) (method string) {
	var builder strings.Builder
	types := toGolangTypes(append([]*parse.Type{m.RetType}, m.ReqTypes...), false)
	fmt.Fprintf(&builder, "%s(", m.Name)
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
	} else if m.RetType.Kind == parse.TypeKindStream {
		fmt.Fprintf(&builder, "(stream %s, onFinish func(), err error)", types[0])
	} else {
		fmt.Fprintf(&builder, "(%s, error)", types[0])
	}
	return builder.String()
}

func toGolangTypes(ts []*parse.Type, closer bool) []string {
	var types []string
	for _, t := range ts {
		types = append(types, toGolangType(t, closer))
	}
	return types
}

func toGolangType(t *parse.Type, closer bool) string {
	switch t.Kind {
	case parse.TypeKindNormal:
		return t.Name
	case parse.TypeKindMessage:
		return "*" + t.Name
	case parse.TypeKindStream:
		if closer {
			return toGlangMap2[t.Name]
		}
		return toGlangMap1[t.Name]
	default:
		return ""
	}
}
