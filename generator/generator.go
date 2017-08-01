package generator

import (
	"bytes"
	"fmt"
	"go/format"
	"io/ioutil"
	"path/filepath"
	"text/template"

	"github.com/devimteam/microgen/util"
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
	outputDir string
	pkgName   string
}

func NewGenerator(ts []*Template, data interface{}, outputDir string) Generator {
	return &generator{
		templates: ts,
		data:      data,
		outputDir: outputDir,
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

		err = tpl.ExecuteTemplate(buf, filepath.Base(t.TemplatePath), g.data)
		if err != nil {
			return fmt.Errorf("error when execute template engine: %v", err)
		}

		fmtSrc, err := format.Source(buf.Bytes())
		if err != nil {
			return fmt.Errorf("error when fmt source: %v, see code below \n %s", err, buf.Bytes())
		}

		err = ioutil.WriteFile(filepath.Join(g.outputDir, t.ResultPath), fmtSrc, 0777)
		if err != nil {
			return fmt.Errorf("error when write file: %v", err)
		}
	}

	return nil
}
