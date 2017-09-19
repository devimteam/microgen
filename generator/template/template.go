package template

import (
	"github.com/devimteam/microgen/generator/write_strategy"
	"github.com/vetcher/jennifer/jen"
)

type Template interface {
	DefaultPath() string
	ChooseStrategy() (write_strategy.Strategy, error)
	Render() *jen.Statement
}
