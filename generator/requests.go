package generator

import (
	"github.com/cv21/microgen/parser"
	"text/template"
	"os"
	"github.com/cv21/microgen/util"
	"fmt"
)

const REQUEST_TEMPLATE =`
{{range .FuncSignatures}}
	type {{.Name}}Request struct {
		{{range .Params}}
			{{.Name | ToUpperFirst}} {{.Type}}
		{{end}}
	}
{{end}}
`

func GenerateRequestsFile(i *parser.Interface) {
	fm := template.FuncMap{
		"ToUpperFirst": util.ToUpperFirst,
	}
	t := template.New("requests")
	t.Funcs(fm)
	t.Parse(REQUEST_TEMPLATE)
	err := t.Execute(os.Stdout, i)
	fmt.Println("Err:", err)
}