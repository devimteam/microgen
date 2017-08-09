package template

import (
	. "github.com/dave/jennifer/jen"
	"github.com/devimteam/microgen/parser"
)

const (
	MiddlewareTypeName = "Middleware"
)

type MiddlewareTemplate struct {
	PackagePath string
}

// Render middleware decorator
//
//		type Middleware func(svc.StringService) svc.StringService
//
func (t MiddlewareTemplate) Render(i *parser.Interface) *File {
	f := NewFile(i.PackageName)
	f.Type().Id(MiddlewareTypeName).Func().Call(Qual(t.PackagePath, i.Name)).Qual(t.PackagePath, i.Name)
	return f
}

func (MiddlewareTemplate) Path() string {
	return "./middleware/middleware.go"
}
