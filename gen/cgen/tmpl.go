package cgen

import "text/template"

var (
	macroTmpl                  = must(_macroTmpl)
	statementTmpl              = must(_statementTmpl)
	includesTmpl               = must(_includesTmpl)
	structStateTmpl            = must(_structStateTmpl)
	structTmpl                 = must(_structTmpl)
	serviceMethodTmpl          = must(_serviceMethodTmpl)
	sourceFileIncludesTmpl     = must(_sourceFileIncludesTmpl)
	registerServiceTmpl        = must(_registerServiceTmpl)
	argumentInitAndDestroyTmpl = must(_argumentInitAndDestroyTmpl)
	marshalFuncTmpl            = must(_marshalFuncTmpl)
	unmarshalFuncTmpl          = must(_unmarshalFuncTmpl)
	handlerTmpl                = must(_handlerTmpl)
	clientMethodTmpl           = must(_clientMethodTmpl)
	clientCallTmpl             = must(_clientCallTmpl)
)

func must(tmpl string) *template.Template {
	return template.Must(template.New("").Parse(tmpl))
}

const _statementTmpl = `// This is code generated by hgen. DO NOT EDIT!!!
// hgen version: {{.Version}}
// source: {{.Source}}

`

const _includesTmpl = `
{{- range . -}}
#include {{ . }}
{{ end -}}
`

const _structStateTmpl = `
{{- range . }}
struct {{.Name}};
{{- end }}
`

const _structTmpl = `
struct {{ .Name }}{
{{- range .Members }}		
	{{.}};
{{- end }}
};
`

const _serviceMethodTmpl = `
// server should implement following functions for service: {{.ServiceName}}
//**********************************************************
{{- range .Methods }}
{{.}}
{{- end }}
//**********************************************************
void register_{{.ServiceName}}_service(server_t*);

`

const _sourceFileIncludesTmpl = `#include "{{.Header}}"

{{ range .Stdlib -}}
#include <{{.}}>
{{ end }}
{{ range .Includes -}}
#include "{{.}}"
{{ end -}}
`

const _registerServiceTmpl = `

{{ $name:=.Name -}}
void register_{{.Name}}_service(server_t* svr) {
	{{- range .Methods }}
	server_register(svr, "{{$name}}.{{.MethodName}}", {{$name}}_{{.MethodName}}_handler);
	{{- end }}
}`

const _argumentInitAndDestroyTmpl = `
void {{.Name}}_init(struct {{.Name}}* data)
{{- if eq (len .MessageMems) 0 }} {}
{{- else }} {
	{{- range .MessageMems}}
	data->{{.MemName}} = malloc(sizeof(struct {{.MemType.TypeName}}));
	{{.MemType.TypeName}}_init(data->{{.MemName}});
	{{- end }}
}
{{- end }}
void {{.Name}}_destroy(struct {{.Name}}* data)
{{- if eq (len .MessageMems) 0 }} {}
{{- else }} {
	{{- range .MessageMems}}
	{{.MemType.TypeName}}_destroy(data->{{.MemName}});
	free(data->{{.MemName}});
	{{- end }}
}
{{- end -}}
`

const _marshalFuncTmpl = `
cJSON* {{.TypeName}}_marshal(struct {{.TypeName}}* data, error_t* err) {
	cJSON* root = NULL;
	{{ if .MessageMem -}}
	cJSON* item = NULL;
	{{ end }}
	if (data == NULL) goto bad;
    root = cJSON_CreateObject();
    if (!root) goto bad;
	{{- range .Message.Mems -}}
	{{- if eq .MemType.TypeKind 2 }}
    if (data->{{.MemName}} == NULL) {
        if (cJSON_AddNullToObject(root, "{{.MemName}}") == NULL) goto bad;
    } else {
		item = {{ .MemType.TypeName }}_marshal(data->{{.MemName}}, err);
		if (!err->null) goto bad;
    	if (!cJSON_AddItemToObject(root, "{{ .MemName }}", item)) goto bad;
    }
	{{- else if eq .MemType.TypeName "string" }}
    if (cJSON_AddStringToObject(root, "{{ .MemName }}", data->{{.MemName}}) == NULL) goto bad;
	{{- else }}
    if (cJSON_AddNumberToObject(root, "{{ .MemName }}", (double)data->{{.MemName}}) == NULL) goto bad;
	{{- end }}
	{{- end }}
	return root;
bad:
	if (!err->null) MARSHAL_FAILED("{{.TypeName}}")
    if (root) cJSON_Delete(root);
	return NULL;
}
`

const _unmarshalFuncTmpl = `
void {{.TypeName}}_unmarshal(struct {{.TypeName}}* dst, char* data, error_t* err) {
    cJSON* root = NULL;
    cJSON* item = NULL;

    root = cJSON_Parse(data);
    if (!root) goto bad;
	{{- $map:=.IDL2CType -}}
	{{ range .Message.Mems }}
    item = cJSON_GetObjectItemCaseSensitive(root, "{{ .MemName }}");
	{{- if eq .MemType.TypeKind 2 }}
    if (cJSON_IsNull(item))
        dst->{{ .MemName }} = NULL;
    else {
		if (!item || !cJSON_IsObject(item)) goto bad;
    	data = cJSON_Print(item);
		{{.MemType.TypeName}}_unmarshal(dst->{{.MemName}}, data, err);
		if (!err->null) goto bad;
    }
	{{- else if eq .MemType.TypeName "string" }}
	if (!item || !cJSON_IsString(item)) goto bad;
	dst->{{.MemName}} = strdup(cJSON_GetStringValue(item));
	{{- else if or (eq .MemType.TypeName "float32") (eq .MemType.TypeName "float64")}}
	if (!item || !cJSON_IsNumber(item)) goto bad;
	dst->{{.MemName}} = ({{index $map .MemType.TypeName}})item->valuedouble;
	{{- else }}
	if (!item || !cJSON_IsNumber(item)) goto bad;
	dst->{{.MemName}} = ({{index $map .MemType.TypeName}})item->valueint;
	{{- end }}
	{{- end }}
    cJSON_Delete(root);
    return;
bad:
    if (!err->null) UNMARSHAL_FAILED("{{.TypeName}}");
    if (root) cJSON_Delete(root);
}
`

const _handlerTmpl = `
void {{.FuncName}}_handler(request_t* req, error_t* err, struct argument* resp) {
	{{- range .Defines }}
	{{.}}
	{{- end }}
	{{- if .MessageResp }}
	cJSON* root = NULL;
	{{- end }}

	{{ .ArgChecks }}
	{{- range .ArgInits }}
	{{.}}
	{{- end }}
	{{- range .ArgUnmarshals }}
	{{.}}
	{{- end }}
	
	{{if not .NoResp}}res = {{end}}{{.FuncName}}({{.CallArgs}});
	if (!err->null) goto end;
	{{.Resp}}
end:{{.End}}
	return;
}
`

const _clientMethodTmpl = `
{{- range . }}
{{.}}
{{- end }}
`

const _clientCallTmpl = `
{{ .FuncSignature }} {
	int free_data = 1;
    struct client_request req;
    struct argument resp;
	error_t* err = &client->err;
	{{- if ne (len .MessageArgs) 0 }}
	char* data = NULL;
	{{- end }}
	{{- range .MessageArgs }}
    cJSON* {{.}} = NULL;
	{{- end }}
	{{ .RespDefine }}

	{{ .RequestInit }}
	{{- range .ArgInits }}
	{{ . }}
	{{- end }}
	client_call(client, &req, &resp);
    if (!client->err.null) goto end;
	{{ .RespCheck }}
	{{- .RespUnmarshal }}
end:
    if (resp.data && free_data) free(resp.data);
    if (resp.type_name) free(resp.type_name);
	{{- range .MessageArgs }}
    if ({{.}}) cJSON_Delete({{.}});
	{{- end }}
    return v;
}
`

const _macroTmpl = `
#define invalid_argcnt(err, want, got) \
    errorf(err, "expected count of arugments is %d, but got %d", want, got)
#define invalid_type(err, want, got) \
    errorf(err, "expected argument type: %s, but got %s", want, got)
#define invalid_type_size(err, t, want, got) \
    errorf(err, "expected size for type %s is %d, but got %d", t, want, got)
#define CHECK_ARG_CNT(want, got)                   \
    {                                              \
        if (got != want) {                         \
            invalid_argcnt(err, got, req->argcnt); \
			{{.}};                                \
        }                                          \
    }
#define CHECK_ARG_TYPE(want, got)         \
    {                                     \
        if (strcmp(want, got) != 0) {     \
            invalid_type(err, want, got); \
			{{.}};                       \
        }                                 \
    }
#define CHECK_ARG_SIZE(t, want, got)              \
    {                                             \
        if (want != got) {                        \
            invalid_type_size(err, t, want, got); \
			{{.}};                               \
        }                                         \
    }
#define MARSHAL_FAILED(obj) error_put(err, "marshal struct " obj " failed");
#define UNMARSHAL_FAILED(obj) error_put(err, "unmarshal struct " obj " failed")`
