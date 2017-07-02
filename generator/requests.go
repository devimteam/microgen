package generator

import (
	"github.com/cv21/microgen/parser"
	"text/template"
	"os"
	"strings"
	"fmt"
)

const REQUEST_TEMPLATE =`
{{range .}}
	type {{.Name}}Request struct {
		{{range .Params}}
			{{.Name | ToUpper}} {{.Type}}
		{{end}}
	}
{{end}}
`

func GenerateRequestsFile(fs []*parser.FuncSignature) {
	fm := template.FuncMap{
		"ToUpper": strings.ToUpper,
	}
	t := template.New("requests")
	t.Funcs(fm)
	t.Parse(REQUEST_TEMPLATE)
	t.Execute(os.Stdout, fs)
}