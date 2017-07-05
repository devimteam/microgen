package generator

import (
	"os"
	"text/template"
	"github.com/cv21/microgen/util"
	"fmt"
	"io/ioutil"
)

type Template interface {
	Render(data interface{}) error
}

type Generator interface {
	Generate() error
}

type CommonTemplate struct {
	Name string
	FileName string
	Template string
	TemplatePath string
}

type generator struct {
	templates []Template
	data interface{}
}

func NewGenerator(ts []Template, data interface{}) Generator {
	return &generator{
		templates: ts,
		data: data,
	}
}

func (t CommonTemplate) Render(data interface{}) error {
	fm := template.FuncMap{
		"ToUpperFirst": util.ToUpperFirst,
		"ToSnakeCase": util.ToSnakeCase,
	}

	tmpl := template.New(t.Name).Funcs(fm)

	if t.TemplatePath != "" {
		b, err := ioutil.ReadFile(t.TemplatePath)
		if err != nil {
			return fmt.Errorf("error when load template file: %v", err)
		}

		tmpl, err = tmpl.Parse(string(b))
		if err != nil {
			return fmt.Errorf("error when parse template from file: %v", err)
		}
	} else {
		var err error
		tmpl, err = tmpl.Parse(t.Template)
		if err != nil {
			return fmt.Errorf("error when parse template: %v", err)
		}
	}

	f, err := os.Create(t.FileName)
	if err != nil {
		return fmt.Errorf("error when create file for template: %v", err)
	}

	err = tmpl.Execute(f, data)
	if err != nil {
		return fmt.Errorf("error when execute template engine: %v", err)
	}

	return f.Close()
}

func (g *generator) Generate() error {
	for _, t := range g.templates {
		err := t.Render(g.data)
		if err != nil {
			return fmt.Errorf("error when render template: %v", err)
		}
	}

	return nil
}