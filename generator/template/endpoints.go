package template

import (
	. "github.com/dave/jennifer/jen"
	"github.com/devimteam/microgen/parser"
)

type EndpointsTemplate struct {
}

func (EndpointsTemplate) Render(i *parser.Interface) *File {
	f := NewFile(i.PackageName)

	f.Type().Id("Endpoints").StructFunc(func(g *Group) {
		for _, signature := range i.FuncSignatures {
			g.Id(signature.Name+"Endpoint").Qual("github.com/go-kit/kit/endpoint", "Endpoint")
		}
	})

	for _, signature := range i.FuncSignatures {
		endpointFunc(signature)
	}

	return f
}

func (EndpointsTemplate) Path() string {
	return "./endpoints.go"
}

func endpointFunc(signature *parser.FuncSignature) Code {
	return Func().
		Params(Id("e").Op("*").Id("Endpoints")).
		Id(signature.Name).
		Params(funcParams(signature.Params)).
		Params(funcParams(signature.Results)).
		BlockFunc(
			func(g *Group) {
				Return()
				//g.Id("req").Op(":=").Id(signature.Name+"Request").StructFunc(func(g *Group) {
				//
				//})
			})
}
