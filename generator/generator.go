package generator

import (
	"fmt"

	"github.com/devimteam/microgen/generator/template"
	"github.com/devimteam/microgen/generator/write_strategy"
)

type Generator interface {
	Generate() error
}

type generationUnit struct {
	template template.Template

	writeMethod write_strategy.Strategy
	absOutPath  string
}

func NewGenUnit(tmpl template.Template, outPath string) (*generationUnit, error) {
	err := tmpl.Prepare()
	if err != nil {
		return nil, fmt.Errorf("%s: prepare error: %v", tmpl.DefaultPath(), err)
	}
	strategy, err := tmpl.ChooseStrategy()
	if err != nil {
		return nil, err
	}
	return &generationUnit{
		template:    tmpl,
		absOutPath:  outPath,
		writeMethod: strategy,
	}, nil
}

func (g *generationUnit) Generate() error {
	code := g.template.Render()
	err := g.writeMethod.Write(code)
	if err != nil {
		return fmt.Errorf("write error: %v", err)
	}
	return nil
}
