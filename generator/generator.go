package generator

import (
	"bytes"
	"fmt"
	"go/format"
	"io/ioutil"
	"path/filepath"
	"text/template"

	"github.com/cv21/microgen/util"
)

type Generator interface {
	Generate() error
}

type Template struct {
	TemplatePath string
	ResultPath   string
}

type generator struct {
	templates []*Template
	data      interface{}
}

func NewGenerator(ts []*Template, data interface{}) Generator {
	return &generator{
		templates: ts,
		data:      data,
	}
}

func (g *generator) Generate() error {
	var templateFiles []string

	for _, t := range g.templates {
		templateFiles = append(templateFiles, t.TemplatePath)
	}

	fm := template.FuncMap{
		"ToUpperFirst": util.ToUpperFirst,
		"ToSnakeCase":  util.ToSnakeCase,
	}

	tpl, err := template.New("main").Funcs(fm).ParseFiles(templateFiles...)
	if err != nil {
		return fmt.Errorf("error when parse files: %v", err)
	}

	for _, t := range g.templates {
		buf := bytes.NewBuffer(nil)

		//f, err := os.Create(t.ResultPath)
		//if err != nil {
		//	return fmt.Errorf("error when create file for template: %v", err)
		//}

		err = tpl.ExecuteTemplate(buf, filepath.Base(t.TemplatePath), g.data)
		if err != nil {
			return fmt.Errorf("error when execute template engine: %v", err)
		}

		fmtSrc, err := format.Source(buf.Bytes())
		if err != nil {
			return fmt.Errorf("error when fmt source: %v", err)
		}

		err = ioutil.WriteFile(t.ResultPath, fmtSrc, 0777)
		if err != nil {
			return fmt.Errorf("error when write file: %v", err)
		}

		//err = f.Close()
		//if err != nil {
		//	return fmt.Errorf("error when close result file: %v", err)
		//}
	}

	return nil
}
