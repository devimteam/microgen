package generator

import (
	"context"
	"errors"
	"fmt"

	"github.com/devimteam/microgen/generator/template"
	"github.com/devimteam/microgen/generator/write_strategy"
)

const (
	Version           = "1.0.0-alpha"
	defaultFileHeader = `Code generated by "microgen ` + Version + `"; DO NOT EDIT.`
)

var (
	EmptyTemplateError = errors.New("empty template")
	EmptyStrategyError = errors.New("empty strategy")
)

type Generator interface {
	Generate() error
}

type generationUnit struct {
	template template.Template

	writeStrategy write_strategy.Strategy
	absOutPath    string
}

func NewGenUnit(ctx context.Context, tmpl template.Template, outPath string) (*generationUnit, error) {
	err := tmpl.Prepare(ctx)
	if err != nil {
		return nil, fmt.Errorf("%s: prepare error: %v", tmpl.DefaultPath(), err)
	}
	strategy, err := tmpl.ChooseStrategy(ctx)
	if err != nil {
		return nil, err
	}
	return &generationUnit{
		template:      tmpl,
		absOutPath:    outPath,
		writeStrategy: strategy,
	}, nil
}

func (g *generationUnit) Generate(ctx context.Context) error {
	if g.template == nil {
		return EmptyTemplateError
	}
	if g.writeStrategy == nil {
		return EmptyStrategyError
	}
	code := g.template.Render(ctx)
	err := g.writeStrategy.Write(code)
	if err != nil {
		return fmt.Errorf("write error: %v", err)
	}
	return nil
}
