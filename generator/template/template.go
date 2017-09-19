package template

import (
	"github.com/devimteam/microgen/generator/write_method"
	"github.com/vetcher/jennifer/jen"
)

type Template interface {
	DefaultPath() string
	ChooseMethod() (write_method.Method, error)
	Render() *jen.Statement
}
