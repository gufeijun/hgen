package nodegen

import (
	"fmt"
	"gufeijun/hustgen/gen/utils"
	"gufeijun/hustgen/parse"
	"strings"
)

var unmarshalMap = map[string]string{
	"int8":    "readInt8",
	"int16":   "readInt16LE",
	"int32":   "readInt32LE",
	"int64":   "readBigInt64LE",
	"uint8":   "readUInt8",
	"uint16":  "readUInt16LE",
	"uint32":  "readUInt32LE",
	"uint64":  "readBigUInt64LE",
	"float32": "readFloatLE",
	"float64": "readDoubleLE",
}

var marshalMap = map[string]string{
	"int8":    "writeInt8",
	"int16":   "writeInt16LE",
	"int32":   "writeInt32LE",
	"int64":   "writeBigInt64LE",
	"uint8":   "writeUInt8",
	"uint16":  "writeUInt16LE",
	"uint32":  "writeUInt32LE",
	"uint64":  "writeBigUInt64LE",
	"float32": "writeFloatLE",
	"float64": "writeDoubleLE",
}

func buildNodeMethod(method *parse.Method) *methodDesc {
	var desc strings.Builder
	var signature strings.Builder
	fmt.Fprintf(&signature, "%s(", method.Name)
	for i, t := range method.ReqTypes {
		if t.Name == "void" {
			break
		}
		fmt.Fprintf(&signature, "arg%d", i+1)
		if i != len(method.ReqTypes)-1 {
			signature.WriteString(", ")
		}
		fmt.Fprintf(&desc, "\n\t// arg%d: %s", i+1, t.Name)
	}
	signature.WriteByte(')')
	if name := method.RetType.Name; name != "void" {
		fmt.Fprintf(&desc, "\n\t// ret:  %s", name)
	}
	return &methodDesc{
		Desc:      desc.String(),
		Signature: signature.String(),
	}
}

func buildMethodsName(s *parse.Service) (string, []string) {
	var builder strings.Builder
	var methods []string
	for i, method := range s.Methods {
		fmt.Fprintf(&builder, `"%s"`, method.Name)
		methods = append(methods, method.Name)
		if i != len(s.Methods)-1 {
			fmt.Fprint(&builder, ", ")
		}
	}
	return builder.String(), methods
}

func buildChecks(method *parse.Method) []string {
	var builder strings.Builder
	var checks []string
	for i, t := range method.ReqTypes {
		fmt.Fprintf(&builder, `if (args[%d].name != "%s"`, i, t.Name)
		if t.Kind == parse.TypeKindNormal && t.Name != "string" {
			fmt.Fprintf(&builder, ` || args[%d].data.length != %d`, i, utils.TypeLength[t.Name])
		}
		fmt.Fprintf(&builder, `) throw "invalid type";`)
		checks = append(checks, builder.String())
		builder.Reset()
	}
	return checks
}

func buildUnmarshalArgs(method *parse.Method) []string {
	var data []string
	var builder strings.Builder
	for i, t := range method.ReqTypes {
		if t.Name == "string" {
			fmt.Fprintf(&builder, `let arg%d = args[%d].data.toString();`, i, i)
		} else if t.Kind == parse.TypeKindMessage {
			fmt.Fprintf(&builder, `let arg%d = JSON.parse(args[%d].data.toString());`, i, i)
		} else {
			fmt.Fprintf(&builder, `let arg%d = Number(args[%d].data.%s());`, i, i, unmarshalMap[t.Name])
		}
		data = append(data, builder.String())
		builder.Reset()
	}
	return data
}

func buildCallHandler(method *parse.Method) string {
	var args string
	for i, _ := range method.ReqTypes {
		args += fmt.Sprintf("arg%d", i)
		if i != len(method.ReqTypes)-1 {
			args += ", "
		}
	}
	if method.RetType.Name == "void" {
		return fmt.Sprintf("await impl.%s(%s);", method.Name, args)
	}
	return fmt.Sprintf("let res = await impl.%s(%s);", method.Name, args)
}

func buildRespDesc(method *parse.Method) *respDesc {
	t := method.RetType
	if t.Name == "void" {
		return &respDesc{
			TypeKind: parse.TypeKindNoRTN,
			Name:     "",
			Data:     `""`,
		}
	}
	if t.Name == "string" {
		return &respDesc{
			TypeKind: parse.TypeKindNormal,
			Name:     t.Name,
			Data:     "res",
		}
	}
	if t.Kind == parse.TypeKindMessage {
		return &respDesc{
			TypeKind: parse.TypeKindMessage,
			Name:     t.Name,
			Data:     "JSON.stringify(res)",
		}
	}
	var builder strings.Builder
	fmt.Fprintf(&builder, "let data = Buffer.alloc(%d);", utils.TypeLength[t.Name])
	src := "res"
	if t.Name == "int64" || t.Name == "uint64" {
		src = "BigInt(res)"
	}
	fmt.Fprintf(&builder, "\n\t\tdata.%s(%s);", marshalMap[t.Name], src)
	return &respDesc{
		Prepare:  builder.String(),
		TypeKind: parse.TypeKindNormal,
		Name:     t.Name,
		Data:     "data",
	}
}

func buildMarshalArg(i int, t *parse.Type) string {
	const format = `req.args.push({
            typeKind: %d,
            name: '%s',
            data: %s,
        })`
	if t.Name == "string" {
		return fmt.Sprintf(format, t.Kind, t.Name, fmt.Sprintf("arg%d", i))
	}
	if t.Kind == parse.TypeKindMessage {
		return fmt.Sprintf(format, t.Kind, t.Name, fmt.Sprintf("JSON.stringify(arg%d)", i))
	}
	var builder strings.Builder
	fmt.Fprintf(&builder, "let buf%d = Buffer.alloc(%d)\n", i, utils.TypeLength[t.Name])
	src := fmt.Sprintf("arg%d", i)
	if t.Name == "uint64" || t.Name == "int64" {
		src = fmt.Sprintf("BigInt(%s)", src)
	}
	fmt.Fprintf(&builder, "\t\tbuf%d.%s(%s)\n\t\t", i, marshalMap[t.Name], src)
	fmt.Fprintf(&builder, format, t.Kind, t.Name, fmt.Sprintf("buf%d", i))
	return builder.String()
}

func buildRespCheck(method *parse.Method) string {
	var builder strings.Builder
	ret := method.RetType
	if ret.Name != "void" {
		fmt.Fprintf(&builder, `resp.name != "%s"`, ret.Name)
	}
	if ret.Kind == parse.TypeKindNormal && ret.Name != "string" {
		if ret.Name != "void" {
			fmt.Fprintf(&builder, " || ")
		}
		fmt.Fprintf(&builder, `resp.dataLen != %d`, utils.TypeLength[ret.Name])
	}
	return builder.String()
}

func buildUnmashalResp(t *parse.Type) string {
	if t.Name == "string" {
		return "resolve(resp.data.toString());"
	}
	if t.Name == "void" {
		return "resolve();"
	}
	if t.Kind == parse.TypeKindMessage {
		return "resolve(JSON.parse(resp.data.toString()));"
	}
	return fmt.Sprintf(`resolve(Number(resp.data.%s()));`, unmarshalMap[t.Name])
}

func buildClientMethod(method *parse.Method) *clientMethod {
	data := new(clientMethod)
	data.Service = method.Service.Name
	data.Name = method.Name
	data.MethodDesc = buildNodeMethod(method)
	data.ArgCnt = len(method.ReqTypes)
	for i, t := range method.ReqTypes {
		data.MashalArgs = append(data.MashalArgs, buildMarshalArg(i+1, t))
	}
	data.RespCheck = buildRespCheck(method)
	data.UnmashalResp = buildUnmashalResp(method.RetType)

	return data
}
