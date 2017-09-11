package generator

import (
	"github.com/devimteam/microgen/generator/template"
	"github.com/vetcher/jennifer/jen"
)

type Template interface {
	Render(data *template.GenerationInfo) *jen.Statement
}
