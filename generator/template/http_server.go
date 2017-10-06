package template

import (
	"path/filepath"

	. "github.com/dave/jennifer/jen"
	"github.com/devimteam/microgen/generator/write_strategy"
	"github.com/devimteam/microgen/util"
)

type httpServerTemplate struct {
	Info *GenerationInfo
}

func NewHttpServerTemplate(info *GenerationInfo) Template {
	return &httpServerTemplate{
		Info: info,
	}
}

func (t *httpServerTemplate) DefaultPath() string {
	return "./transport/http/server.go"
}

func (t *httpServerTemplate) ChooseStrategy() (write_strategy.Strategy, error) {
	return write_strategy.NewCreateFileStrategy(t.Info.AbsOutPath, t.DefaultPath()), nil
}

func (t *httpServerTemplate) Prepare() error {
	return nil
}

func (t *httpServerTemplate) Render() write_strategy.Renderer {
	f := NewFile("transporthttp")
	f.PackageComment(FileHeader)
	f.PackageComment(`Please, do not edit.`)

	f.Func().Id("NewHTTPHandler").Params(
		Id("endpoints").Op("*").Qual(t.Info.ServiceImportPath, "Endpoints"),
		Id("opts").Op("...").Qual(PackagePathGoKitTransportHTTP, "ServerOption"),
	).Params(
		Qual(PackagePathHttp, "Handler"),
	).BlockFunc(func(g *Group) {
		g.Id("handler").Op(":=").Qual(PackagePathHttp, "NewServeMux").Call()
		for _, fn := range t.Info.Iface.Methods {
			g.Id("handler").Dot("Handle").Call(
				Lit("/"+util.ToSnakeCase(fn.Name)),
				Qual(PackagePathGoKitTransportHTTP, "NewServer").Call(
					Line().Id("endpoints").Dot(endpointStructName(fn.Name)),
					Line().Qual(pathToHttpConverter(t.Info.ServiceImportPath), httpDecodeRequestName(fn)),
					Line().Qual(pathToHttpConverter(t.Info.ServiceImportPath), httpEncodeResponseName(fn)),
					Line().Id("opts").Op("..."),
				),
			)
		}
	})

	return f
}

func pathToHttpConverter(servicePath string) string {
	return filepath.Join(servicePath, "transport/converter/http")
}
