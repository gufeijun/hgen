package nodegen

import "text/template"

var (
	statementTmpl        = must(_statementTmpl)
	serviceInterfaceTmpl = must(_serviceInterfaceTmpl)
	registerServiceTmpl  = must(_registerServiceTmpl)
	moduleExportsTmpl    = must(_moduleExportsTmpl)
)

func must(tmpl string) *template.Template {
	return template.Must(template.New("").Parse(tmpl))
}

const _statementTmpl = `// This is code generated by hgen. DO NOT EDIT!!!
// hgen version: {{.Version}}
// source: {{.Source}}

`

const _serviceInterfaceTmpl = `
class {{.Name}}Interface {
	{{- range .Methods -}}	
	{{- .Desc }}
	async {{.Signature}} {
		throw "No implementation";
	}
	{{- end }}
};
`
const checkImplementsTmpl = `
function checkImplements(impl, service, methods) {
    methods.forEach(method => {
        if (impl[method] == undefined)
            throw ` + "`should implement method ${method} for service ${service}`;" + `
    })
}
`

const _registerServiceTmpl = `
function register{{.Name}}Service(svr, impl) {
	{{- $name:= .Name }}
	checkImplements(impl, "{{.Name}}", [{{.MethodsName}}]);
	svr.register(svr, {
		name: "{{.Name}}",
		methods: {
			{{- range .Methods}}
			{{.}}: {{$name}}{{.}}Handler(impl),
			{{- end}}
		}
	});
}
`

const _moduleExportsTmpl = `
module.exports = {
{{- range . }}
	register{{.}}Service,
	{{.}}Interface,
{{- end }}
}
`