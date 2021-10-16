package cgen

import "text/template"

var (
	statementTmpl              = must(_statementTmpl)
	includesTmpl               = must(_includesTmpl)
	structStateTmpl            = must(_structStateTmpl)
	structTmpl                 = must(_structTmpl)
	serviceMethodTmpl          = must(_serviceMethodTmpl)
	sourceFileIncludesTmpl     = must(_sourceFileIncludesTmpl)
	registerServiceTmpl        = must(_registerServiceTmpl)
	argumentInitAndDestroyTmpl = must(_argumentInitAndDestroyTmpl)
	handlerTmpl                = must(_handlerTmpl)
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
void register_{{.Name}}_service(server_t* svr){
	{{- range .Methods }}
	server_register(svr, "{{$name}}.{{.MethodName}}", {{$name}}_{{.MethodName}}_handler);
	{{- end }}
}`

const _argumentInitAndDestroyTmpl = `
void {{.Name}}_init(struct {{.Name}}* data)
{{- if eq (len .MessageMems) 0 -}} {}
{{- else -}}
{
	{{- range .MessageMems}}
	data->{{.MemName}} = malloc(sizeof(struct {{.MemType.TypeName}}));
	{{.MemType.TypeName}}_init(data->{{.MemName}});
	{{- end }}
}
{{- end }}
void {{.Name}}_destroy(struct {{.Name}}* data)
{{- if eq (len .MessageMems) 0 -}} {}
{{- else -}}
{
	{{- range .MessageMems}}
	{{.MemType.TypeName}}_destroy(data->{{.MemName}});
	{{- end }}
}
{{- end -}}
`

const _handlerTmpl = `

`
