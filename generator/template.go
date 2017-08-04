package generator

import (
	"github.com/dave/jennifer/jen"
	"github.com/devimteam/microgen/parser"
)

type Template interface {
	Path() string
	Render(data *parser.Interface) *jen.File
}
