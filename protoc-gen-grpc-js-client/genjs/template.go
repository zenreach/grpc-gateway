package genjs

import (
	"bytes"
	"strings"
	"text/template"

	"github.com/golang/glog"
	"github.com/grpc-ecosystem/grpc-gateway/protoc-gen-grpc-gateway/descriptor"
)

type binding struct {
	*descriptor.Binding
	URL        string
	DeleteVars []string
	BodyVar    string
	ParamVar   string
}

const paramVar = "p"

func jsGetter(fieldPath string) string {
	js := paramVar
	for _, f := range strings.Split(fieldPath, ".") {
		js += ("['" + f + "']")
	}
	return js
}

func urlize(path string) (string, []string) {
	if path == "" {
		return "", []string{}
	}

	const (
		init = iota
		field
		nested
		restart
	)
	var st = init

	deleteList := make([]string, 0)
	var url, del bytes.Buffer
	url.WriteString(`'`)
	for _, r := range path {
		switch st {
		case init:
			if r == '{' {
				url.WriteString(`' + ` + paramVar + `['`)
				del.WriteString(paramVar + `['`)
				st = field
			} else {
				url.WriteRune(r)
			}
		case field:
			if r == '.' {
				url.WriteString(`']['`)
				del.WriteString(`']['`)
			} else if r == '=' {
				url.WriteString(`']`)
				del.WriteString(`']`)
				deleteList = append(deleteList, del.String())
				del.Reset()
				st = nested
			} else if r == '}' {
				url.WriteString(`']`)
				del.WriteString(`']`)
				deleteList = append(deleteList, del.String())
				del.Reset()
				st = restart
			} else {
				url.WriteRune(r)
				del.WriteRune(r)
			}
		case nested:
			if r == '}' {
				st = restart
			}
		case restart:
			if r == '/' {
				url.WriteString(` + '/`)
				st = init
			}
		}
	}
	if st == init {
		url.WriteString(`'`)
	}
	return url.String(), deleteList
}

func (g *generator) applyTemplate(file *descriptor.File) (string, error) {
	w := bytes.NewBuffer(nil)
	if err := headerTemplate.Execute(w, file); err != nil {
		return "", err
	}
	var methodSeen bool
	for _, svc := range file.Services {
		for _, meth := range svc.Methods {
			glog.V(2).Infof("Processing %s.%s", svc.GetName(), meth.GetName())
			methodSeen = true
			for _, b := range meth.Bindings {
				url, delVars := urlize(strings.TrimPrefix(b.PathTmpl.Template, g.reg.Prefix))
				bind := binding{
					Binding:    b,
					URL:        url,
					DeleteVars: delVars,
					ParamVar:   paramVar,
				}
				if b.Body == nil {
					bind.BodyVar = "null"
				} else if b.Body.FieldPath.String() == "" {
					bind.BodyVar = paramVar
				} else {
					bind.BodyVar = jsGetter(b.Body.FieldPath.String())
				}
				if err := handlerTemplate.Execute(w, bind); err != nil {
					return "", err
				}
				break
			}
		}
	}
	if !methodSeen {
		return "", errNoTargetService
	}
	if err := trailerTemplate.Execute(w, file.Services); err != nil {
		return "", err
	}
	return w.String(), nil
}

var (
	headerTemplate = template.Must(template.New("header").Parse(`const xhr = require('../lib/xhr.js')
`))

	handlerTemplate = template.Must(template.New("handler").Parse(`
func {{.Method.Service.GetName}}{{.Method.GetName}}({{.ParamVar}}, conf) {
	url = {{.URL}}
	{{- range $delVar := .DeleteVars}}
	delete {{$delVar}}
	{{- end}}
	{{- if .Body}}
	{{- if .Body.FieldPath}}
	body = {{.BodyVar}}
	delete {{.BodyVar}}
	return xhr(url, '{{.HTTPMethod}}', conf, {{.ParamVar}}, {{.BodyVar}});
	{{- else}}
	return xhr(url, '{{.HTTPMethod}}', conf, null, {{.ParamVar}});
	{{- end}}
	{{- else}}
	return xhr(url, '{{.HTTPMethod}}', conf, {{.ParamVar}});
	{{- end}}
}
`))

	trailerTemplate = template.Must(template.New("trailer").Parse(`
module.exports = {
{{- range $svc_idx, $svc := . -}}
  {{if $svc_idx}},{{end}}
  {{$svc.GetName}}: {
    {{- range $m_idx, $m := $svc.Methods -}}
      {{if $m_idx}},{{end}}
      {{$m.GetName}}: {{$svc.GetName}}{{$m.GetName}}
    {{- end}}
  }
{{- end}}
}`))
)
