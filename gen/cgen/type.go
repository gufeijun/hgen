package cgen

import (
	"fmt"
	"gufeijun/hustgen/gen/utils"
	"gufeijun/hustgen/parse"
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

func toClangType(t *parse.Type, pointer bool) string {
	switch t.Kind {
	case parse.TypeKindNormal:
		return IDLtoCType[t.Name]
	case parse.TypeKindMessage:
		if pointer {
			return fmt.Sprintf("struct %s*", t.Name)
		} else {
			return fmt.Sprintf("struct %s", t.Name)
		}
	default:
	}
	return ""
}

func buildMethod(method *parse.Method, lastArg string) string {
	var builder strings.Builder
	builder.WriteString(toClangType(method.RetType, true))
	fmt.Fprintf(&builder, " %s_%s(", method.Service.Name, method.Name)
	for _, t := range method.ReqTypes {
		builder.WriteString(toClangType(t, true))
		builder.WriteString(", ")
	}
	builder.WriteString(lastArg)
	builder.WriteString(");")
	return builder.String()
}

func buildArgDefines(method *parse.Method) []string {
	defines := make([]string, 0, len(method.ReqTypes)+1)
	for i, t := range method.ReqTypes {
		defines = append(defines, fmt.Sprintf("%s arg%d;", toClangType(t, false), i+1))
	}
	var str string
	t := method.RetType
	if t.Kind == parse.TypeKindMessage || t.Name == "string" {
		str = fmt.Sprintf("%s res = NULL;", toClangType(t, true))
	} else {
		if t.Name == "void" {
			return defines
		}
		str = fmt.Sprintf("%s res;", toClangType(t, true))
	}
	defines = append(defines, str)
	return defines
}

func buildArgInits(method *parse.Method) []string {
	strs := []string{}
	for i, v := range method.ReqTypes {
		if v.Kind == parse.TypeKindMessage {
			strs = append(strs, fmt.Sprintf("%s_init(&arg%d);", v.Name, i+1))
		}
	}
	return strs
}

func buildArgUnmarshals(method *parse.Method) []string {
	strs := make([]string, 0, len(method.ReqTypes))
	for i, t := range method.ReqTypes {
		if t.Kind == parse.TypeKindMessage {
			var builder strings.Builder
			builder.WriteString(fmt.Sprintf("%s_unmarshal(&arg%d, req->args[%d].data, err);", t.Name, i+1, i))
			builder.WriteString("\n\tif (!err->null) goto end;")
			strs = append(strs, builder.String())
			continue
		}
		if t.Name == "string" {
			strs = append(strs, fmt.Sprintf("arg%d = req->args[%d].data;", i+1, i))
			continue
		}
		strs = append(strs, fmt.Sprintf("arg%d = *(%s*)req->args[%d].data;", i+1, IDLtoCType[t.Name], i))
	}
	return strs
}

func buildArgChecks(method *parse.Method) string {
	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("CHECK_ARG_CNT(%d, req->argcnt)", len(method.ReqTypes)))
	for i, t := range method.ReqTypes {
		builder.WriteString("\n\t")
		str := fmt.Sprintf(`CHECK_ARG_TYPE("%s", req->args[%d].type_name)`, t.Name, i)
		builder.WriteString(str)
		builder.WriteString("\n\t")
		if t.Kind == parse.TypeKindMessage || t.Name == "string" {
			continue
		}
		str = fmt.Sprintf(`CHECK_ARG_SIZE("%s", %d, req->args[%d].data_len)`, t.Name, utils.TypeLength[t.Name], i)
		builder.WriteString(str)
	}
	return builder.String()
}

func buildCallArgs(method *parse.Method) string {
	var str string
	for i, t := range method.ReqTypes {
		if t.Kind == parse.TypeKindMessage {
			str += "&"
		}
		str += fmt.Sprintf("arg%d, ", i+1)
	}
	str += "err"
	return str
}

func buildResp(method *parse.Method) string {
	var builder strings.Builder
	t := method.RetType
	if t.Kind == parse.TypeKindMessage {
		builder.WriteString(fmt.Sprintf(`root = %s_marshal(res, err);
    if (!err->null) goto end;
    char* data = cJSON_Print(root);
    build_resp(resp, %d, "%s", strlen(data), data);`, t.Name, parse.TypeKindMessage, t.Name))
		return builder.String()
	}
	if t.Name == "string" {
		builder.WriteString(fmt.Sprintf(`build_resp(resp, 0, "string", res == NULL? 0 : strlen(res), res);`))
	} else if t.Name == "void" {
		builder.WriteString(`build_resp(resp, 4, "", 0, NULL);`)
	} else {
		builder.WriteString(fmt.Sprintf(`build_resp(resp, 0, "%s", %d, (char*)&res);`, t.Name, utils.TypeLength[t.Name]))
	}
	return builder.String()
}

func buildEnd(method *parse.Method) string {
	var builder strings.Builder
	for i, t := range method.ReqTypes {
		if t.Kind == parse.TypeKindMessage {
			fmt.Fprintf(&builder, "\n\t%s_destroy(&arg%d);", t.Name, i+1)
		}
	}
	t := method.RetType
	if t.Name == "string" {
		fmt.Fprintf(&builder, "\n\tfree(res);")
	}
	if t.Kind == parse.TypeKindMessage {
		fmt.Fprintf(&builder, "\n\tif (res) %s_destroy(res);", t.Name)
		fmt.Fprintf(&builder, "\n\tfree(res);")
		fmt.Fprintf(&builder, "\n\tif (root) cJSON_Delete(root);")
	}
	return builder.String()
}

func buildCallFuncSignature(method *parse.Method) string {
	var builder strings.Builder
	builder.WriteString(toClangType(method.RetType, true))
	fmt.Fprintf(&builder, " %s_%s(", method.Service.Name, method.Name)
	for i, t := range method.ReqTypes {
		fmt.Fprintf(&builder, "%s arg%d, ", toClangType(t, true), i+1)
	}
	builder.WriteString("client_t* client)")
	return builder.String()
}

func buildRespArgDefine(method *parse.Method) string {
	ret := method.RetType
	if ret.Name == "void" {
		return ""
	}
	var resp string
	if ret.Kind == parse.TypeKindMessage {
		resp = fmt.Sprintf("%s v = NULL;", toClangType(ret, true))
	} else if ret.Name == "string" {
		resp = "char* v = NULL;"
	} else {
		resp = fmt.Sprintf("%s v = 0;", toClangType(ret, false))
	}
	return resp
}

func buildCallRequestInit(method *parse.Method) string {
	return fmt.Sprintf(`client_request_init(&req, "%s", "%s", %d);`, method.Service.Name, method.Name, len(method.ReqTypes))
}

func buildCallArgInits(method *parse.Method) (res []string) {
	var ii int
	for i, t := range method.ReqTypes {
		if t.Kind == parse.TypeKindMessage {
			ii++
			var builder strings.Builder
			fmt.Fprintf(&builder, `node%d = %s_marshal(arg%d, &client->err);`, ii, t.Name, i+1)
			fmt.Fprint(&builder, "\n\t")
			fmt.Fprint(&builder, `if (client_failed(client)) return`)
			if method.RetType.Name == "void" {
				fmt.Fprint(&builder, ";")
			} else {
				fmt.Fprint(&builder, " v;")
			}
			fmt.Fprint(&builder, "\n\t")
			fmt.Fprintf(&builder, `data = cJSON_Print(node%d);`, ii)
			fmt.Fprint(&builder, "\n\t")
			fmt.Fprintf(&builder, `argument_init_with_option(req.args + %d, %d, "%s", data, strlen(data));`, i, t.Kind, t.Name)
			res = append(res, builder.String())
		}
		var str string
		if t.Name == "string" {
			str = fmt.Sprintf(`arg%d = arg%d == NULL ? "" : arg%d;
	argument_init_with_option(req.args + %d, %d, "%s", arg%d, strlen(arg%d));`, i+1, i+1, i+1,
				i, t.Kind, t.Name, i+1, i+1)
		} else if t.Kind == parse.TypeKindNormal {
			str = fmt.Sprintf(`argument_init_with_option(req.args + %d, %d, "%s", &arg%d, %d);`,
				i, t.Kind, t.Name, i+1, utils.TypeLength[t.Name])
		}
		res = append(res, str)
	}
	return
}

func buildRespCheck(method *parse.Method) string {
	var builder strings.Builder
	ret := method.RetType
	if ret.Kind == parse.TypeKindNoRTN {
		return ""
	}
	fmt.Fprintf(&builder, `CHECK_ARG_TYPE("%s", resp.type_name)`, ret.Name)
	builder.WriteByte('\n')
	if ret.Kind == parse.TypeKindNormal && ret.Name != "string" {
		fmt.Fprintf(&builder, `	CHECK_ARG_SIZE("%s", %d, resp.data_len)`, ret.Name, utils.TypeLength[ret.Name])
		builder.WriteByte('\n')
	}
	return builder.String()
}

func buildRespUnmarshal(method *parse.Method) string {
	ret := method.RetType
	if ret.Name == "string" {
		return `	v = resp.data;
	free_data = 0;`
	}
	if ret.Kind == parse.TypeKindNormal {
		return fmt.Sprintf(`	memcpy(&v, resp.data, %d);`, utils.TypeLength[ret.Name])
	}
	return fmt.Sprintf(`	v = malloc(sizeof(struct %s));
	%s_init(v);
	%s_unmarshal(v, resp.data, &client->err);
	`, ret.Name, ret.Name, ret.Name)
}

func buildAssignment(mem *parse.Member) string {
	if mem.Type.Name == "string" {
		return fmt.Sprintf("dst->%s = src->%s == NULL? NULL : strdup(src->%s);", mem.Name, mem.Name, mem.Name)
	}
	if mem.Type.Kind == parse.TypeKindNormal {
		return fmt.Sprintf("dst->%s = src->%s;", mem.Name, mem.Name)
	}
	return fmt.Sprintf("dst->%s = %s_clone(src->%s);", mem.Name, mem.Type.Name, mem.Name)
}
