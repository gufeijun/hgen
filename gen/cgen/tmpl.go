package cgen

import "text/template"

func must(tmpl string) *template.Template {
	return template.Must(template.New("").Parse(tmpl))
}
