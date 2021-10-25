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
	"void":    "void",
	"string":  "char*",
}

var typeLength = map[string]int{
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

func toClangType(t *service.Type, pointer bool) string {
	switch t.TypeKind {
	case service.TypeKindNormal:
		return IDLtoCType[t.TypeName]
	case service.TypeKindMessage:
		if pointer {
			return fmt.Sprintf("struct %s*", t.TypeName)
		} else {
			return fmt.Sprintf("struct %s", t.TypeName)
		}
	default:
	}
	return ""
}

func buildMethod(method *service.Method, lastArg string) string {
	var builder strings.Builder
	builder.WriteString(toClangType(method.RetType, true))
	fmt.Fprintf(&builder, " %s_%s(", method.Service.Name, method.MethodName)
	for _, t := range method.ReqTypes {
		builder.WriteString(toClangType(t, true))
		builder.WriteString(", ")
	}
	builder.WriteString(lastArg)
	builder.WriteString(");")
	return builder.String()
}

func buildArgDefines(method *service.Method) []string {
	defines := make([]string, 0, len(method.ReqTypes)+1)
	for i, t := range method.ReqTypes {
		defines = append(defines, fmt.Sprintf("%s arg%d;", toClangType(t, false), i+1))
	}
	var str string
	t := method.RetType
	if t.TypeKind == service.TypeKindMessage || t.TypeName == "string" {
		str = fmt.Sprintf("%s res = NULL;", toClangType(t, true))
	} else {
		if t.TypeName == "void" {
			return defines
		}
		str = fmt.Sprintf("%s res;", toClangType(t, true))
	}
	defines = append(defines, str)
	return defines
}

func buildArgInits(method *service.Method) []string {
	strs := []string{}
	for i, v := range method.ReqTypes {
		if v.TypeKind == service.TypeKindMessage {
			strs = append(strs, fmt.Sprintf("%s_init(&arg%d);", v.TypeName, i+1))
		}
	}
	return strs
}

func buildArgUnmarshals(method *service.Method) []string {
	strs := make([]string, 0, len(method.ReqTypes))
	for i, t := range method.ReqTypes {
		if t.TypeKind == service.TypeKindMessage {
			var builder strings.Builder
			builder.WriteString(fmt.Sprintf("%s_unmarshal(&arg%d, req->args[%d].data, err);", t.TypeName, i+1, i))
			builder.WriteString("\n\tif (!err->null) goto end;")
			strs = append(strs, builder.String())
			continue
		}
		if t.TypeName == "string" {
			strs = append(strs, fmt.Sprintf("arg%d = req->args[%d].data;", i+1, i))
			continue
		}
		strs = append(strs, fmt.Sprintf("arg%d = *(%s*)req->args[%d].data;", i+1, IDLtoCType[t.TypeName], i))
	}
	return strs
}

func buildArgChecks(method *service.Method) string {
	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("CHECK_ARG_CNT(%d, req->argcnt)", len(method.ReqTypes)))
	for i, t := range method.ReqTypes {
		builder.WriteString("\n\t")
		str := fmt.Sprintf(`CHECK_ARG_TYPE("%s", req->args[%d].type_name)`, t.TypeName, i)
		builder.WriteString(str)
		builder.WriteString("\n\t")
		if t.TypeKind == service.TypeKindMessage || t.TypeName == "string" {
			continue
		}
		str = fmt.Sprintf(`CHECK_ARG_SIZE("%s", %d, req->args[%d].data_len)`, t.TypeName, typeLength[t.TypeName], i)
		builder.WriteString(str)
	}
	return builder.String()
}

func buildCallArgs(method *service.Method) string {
	var str string
	for i, t := range method.ReqTypes {
		if t.TypeKind == service.TypeKindMessage {
			str += "&"
		}
		str += fmt.Sprintf("arg%d, ", i+1)
	}
	str += "err"
	return str
}

func buildResp(method *service.Method) string {
	var builder strings.Builder
	t := method.RetType
	if t.TypeKind == service.TypeKindMessage {
		builder.WriteString(fmt.Sprintf(`root = %s_marshal(res, err);
    if (!err->null) goto end;
    char* data = cJSON_Print(root);
    build_resp(resp, %d, "%s", strlen(data), data);`, t.TypeName, service.TypeKindMessage, t.TypeName))
		return builder.String()
	}
	if t.TypeName == "string" {
		builder.WriteString(fmt.Sprintf(`build_resp(resp, 0, "string", res == NULL? 0 : strlen(res), res);`))
	} else if t.TypeName == "void" {
		builder.WriteString(`build_resp(resp, 4, "", 0, NULL);`)
	} else {
		builder.WriteString(fmt.Sprintf(`build_resp(resp, 0, "%s", %d, (char*)&res);`, t.TypeName, typeLength[t.TypeName]))
	}
	return builder.String()
}

func buildEnd(method *service.Method) string {
	var builder strings.Builder
	for i, t := range method.ReqTypes {
		if t.TypeKind == service.TypeKindMessage {
			fmt.Fprintf(&builder, "\n\t%s_destroy(&arg%d);", t.TypeName, i+1)
		}
	}
	t := method.RetType
	if t.TypeName == "string" {
		fmt.Fprintf(&builder, "\n\tfree(res);")
	}
	if t.TypeKind == service.TypeKindMessage {
		fmt.Fprintf(&builder, "\n\tif (res) %s_destroy(res);", t.TypeName)
		fmt.Fprintf(&builder, "\n\tfree(res);")
		fmt.Fprintf(&builder, "\n\tif (root) cJSON_Delete(root);")
	}
	return builder.String()
}

func buildCallFuncSignature(method *service.Method) string {
	var builder strings.Builder
	builder.WriteString(toClangType(method.RetType, true))
	fmt.Fprintf(&builder, " %s_%s(", method.Service.Name, method.MethodName)
	for i, t := range method.ReqTypes {
		fmt.Fprintf(&builder, "%s arg%d, ", toClangType(t, true), i+1)
	}
	builder.WriteString("client_t* client)")
	return builder.String()
}

func buildRespArgDefine(method *service.Method) string {
	ret := method.RetType
	if ret.TypeName == "void" {
		return ""
	}
	var resp string
	if ret.TypeKind == service.TypeKindMessage {
		resp = fmt.Sprintf("%s v = NULL;", toClangType(ret, true))
	} else if ret.TypeName == "string" {
		resp = "char* v = NULL;"
	} else {
		resp = fmt.Sprintf("%s v = 0;", toClangType(ret, false))
	}
	return resp
}

func buildCallRequestInit(method *service.Method) string {
	return fmt.Sprintf(`client_request_init(&req, "%s", "%s", %d);`, method.Service.Name, method.MethodName, len(method.ReqTypes))
}

func buildCallArgInits(method *service.Method) (res []string) {
	var ii int
	for i, t := range method.ReqTypes {
		if t.TypeKind == service.TypeKindMessage {
			ii++
			var builder strings.Builder
			str := fmt.Sprintf(`node%d = %s_marshal(arg%d, &client->err);
    if (client_failed(client)) goto end;
    data = cJSON_Print(node%d);`, ii, t.TypeName, i+1, ii)
			builder.WriteString(str)
			str = fmt.Sprintf(`
	argument_init_with_option(req.args + %d, %d, "%s", data, strlen(data));`, i, t.TypeKind, t.TypeName)
			builder.WriteString(str)
			res = append(res, builder.String())
		}
		var str string
		if t.TypeName == "string" {
			str = fmt.Sprintf(`argument_init_with_option(req.args + %d, %d, "%s", arg%d, strlen(arg%d));`,
				i, t.TypeKind, t.TypeName, i+1, i+1)
		} else if t.TypeKind == service.TypeKindNormal {
			str = fmt.Sprintf(`argument_init_with_option(req.args + %d, %d, "%s", &arg%d, %d);`,
				i, t.TypeKind, t.TypeName, i+1, typeLength[t.TypeName])
		}
		res = append(res, str)
	}
	return
}
