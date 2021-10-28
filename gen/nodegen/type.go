package nodegen

import (
	"fmt"
	"gufeijun/hustgen/gen/utils"
	"gufeijun/hustgen/service"
	"strings"
)

var unmarshalMap = map[string]string{
	"int8":   "readInt8",
	"int16":  "readInt16LE",
	"int32":  "readInt32LE",
	"int64":  "readBigInt64LE",
	"uint8":  "readUInt8",
	"uint16": "readUInt16LE",
	"uint32": "readUInt32LE",
	"uint64": "readBigUInt64LE",
}

var marshalMap = map[string]string{
	"int8":   "writeInt8",
	"int16":  "writeInt16LE",
	"int32":  "writeInt32LE",
	"int64":  "writeBigInt64LE",
	"uint8":  "writeUInt8",
	"uint16": "writeUInt16LE",
	"uint32": "writeUInt32LE",
	"uint64": "writeBigUInt64LE",
}

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
		fmt.Fprintf(&desc, "\n\t// ret:  %s", name)
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

func buildChecks(method *service.Method) []string {
	var builder strings.Builder
	var checks []string
	for i, t := range method.ReqTypes {
		fmt.Fprintf(&builder, `if (args[%d].name != "%s"`, i, t.TypeName)
		if t.TypeKind == service.TypeKindNormal && t.TypeName != "string" {
			fmt.Fprintf(&builder, ` || args[%d].data.length != %d`, i, utils.TypeLength[t.TypeName])
		}
		fmt.Fprintf(&builder, `) throw "invalid type";`)
		checks = append(checks, builder.String())
		builder.Reset()
	}
	return checks
}

func buildUnmarshalArgs(method *service.Method) []string {
	var data []string
	var builder strings.Builder
	for i, t := range method.ReqTypes {
		if t.TypeName == "string" {
			fmt.Fprintf(&builder, `let arg%d = args[%d].data;`, i, i)
		} else if t.TypeKind == service.TypeKindMessage {
			fmt.Fprintf(&builder, `let arg%d = JSON.parse(args[%d].data);`, i, i)
		} else {
			fmt.Fprintf(&builder, `let arg%d = Buffer.from(args[%d].data);`, i, i)
			fmt.Fprint(&builder, "\n\t\t")
			fmt.Fprintf(&builder, `arg%d = Number(arg%d.%s());`, i, i, unmarshalMap[t.TypeName])
		}
		data = append(data, builder.String())
		builder.Reset()
	}
	return data
}

func buildCallHandler(method *service.Method) string {
	var args string
	for i, _ := range method.ReqTypes {
		args += fmt.Sprintf("arg%d", i)
		if i != len(method.ReqTypes)-1 {
			args += ", "
		}
	}
	return fmt.Sprintf("let res = await impl.%s(%s);", method.MethodName, args)
}

func buildRespDesc(method *service.Method) *respDesc {
	t := method.RetType
	if t.TypeName == "string" {
		return &respDesc{
			TypeKind: service.TypeKindNormal,
			Name:     t.TypeName,
			Data:     "res",
		}
	}
	if t.TypeKind == service.TypeKindMessage {
		return &respDesc{
			TypeKind: service.TypeKindMessage,
			Name:     t.TypeName,
			Data:     "JSON.stringify(res)",
		}
	}
	var builder strings.Builder
	fmt.Fprintf(&builder, "let data = Buffer.alloc(%d);", utils.TypeLength[t.TypeName])
	src := "res"
	if t.TypeName == "int64" || t.TypeName == "uint64" {
		src = "BigInt(res)"
	}
	fmt.Fprintf(&builder, "\n\t\tdata.%s(%s);", marshalMap[t.TypeName], src)
	return &respDesc{
		Prepare:  builder.String(),
		TypeKind: service.TypeKindNormal,
		Name:     t.TypeName,
		Data:     "data",
	}
}
