package generator

import (
	"github.com/vetcher/godecl/types"
	"github.com/vetcher/jennifer/jen"
)

type Template interface {
	PackageName() string
	Path() string
	Render(data *types.Interface) *jen.Statement
}
