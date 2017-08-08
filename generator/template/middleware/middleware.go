package middleware

import (
	. "github.com/dave/jennifer/jen"
	"github.com/devimteam/microgen/parser"
)

const (
	PackageAliasGoKitEndpoint = "github.com/go-kit/kit/endpoint"
	PackageAliasGoKitLog      = "github.com/go-kit/kit/log"
	PackageAliasTime          = "time"
)

type MiddlewareTemplate struct {
}

func (MiddlewareTemplate) Render(i *parser.Interface) *File {
	f := NewFile(i.PackageName)
	f.Type().Id("Middleware").Func().Call(Qual(PackageAliasGoKitEndpoint, "Endpoint")).Qual(PackageAliasGoKitEndpoint, "Endpoint")
	return f
}

func (MiddlewareTemplate) Path() string {
	return "./middleware/middleware.go"
}
