package template

import (
	. "github.com/dave/jennifer/jen"
	"github.com/devimteam/microgen/parser"
)

const (
	MiddlewareTypeName = "Middleware"
)

type MiddlewareTemplate struct {
}

func (MiddlewareTemplate) Render(i *parser.Interface) *File {
	f := NewFile(i.PackageName)
	f.Type().Id(MiddlewareTypeName).Func().Call(Qual(PackagePathGoKitEndpoint, "Endpoint")).Qual(PackagePathGoKitEndpoint, "Endpoint")
	return f
}

func (MiddlewareTemplate) Path() string {
	return "./middleware/middleware.go"
}
