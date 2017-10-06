package template

import (
	. "github.com/dave/jennifer/jen"
	"github.com/devimteam/microgen/generator/write_strategy"
)

type httpClientTemplate struct {
	Info *GenerationInfo
}

func NewHttpClientTemplate(info *GenerationInfo) Template {
	return &httpClientTemplate{
		Info: info,
	}
}

func (t *httpClientTemplate) DefaultPath() string {
	return "./transport/http/client.go"
}

func (t *httpClientTemplate) ChooseStrategy() (write_strategy.Strategy, error) {
	return write_strategy.NewCreateFileStrategy(t.Info.AbsOutPath, t.DefaultPath()), nil
}

func (t *httpClientTemplate) Prepare() error {
	return nil
}

func (t *httpClientTemplate) Render() write_strategy.Renderer {
	f := NewFile("transporthttp")
	f.PackageComment(FileHeader)
	f.PackageComment(`Please, do not edit.`)

	f.Func().Id("NewHTTPClient").Params(
		Id("addr").Id("string"),
		Id("opts").Op("...").Qual(PackagePathGoKitTransportHTTP, "ClientOption"),
	).Params(
		Qual(t.Info.ServiceImportPath, t.Info.Iface.Name),
		Error(),
	).Block(
		t.clientBody(),
	)

	return f
}

func (t *httpClientTemplate) clientBody() *Statement {
	return If(
		Op("!").Qual(PackagePathStrings, "HasPrefix").Call(Id("addr"), Lit("http")),
	).Block(
		Id("addr").Op("=").Lit("http://").Op("+").Id("addr"),
	).
		Line().List(Id("u"), Err()).Op(":=").Qual(PackagePathUrl, "Parse").Call(Id("addr")).
		Line().If(Err().Op("!=").Nil()).
		Block(
			Return(Nil(), Err()),
		).
		Line().Return(Qual(t.Info.ServiceImportPath, "Endpoints").Values(DictFunc(
		func(d Dict) {
			for _, fn := range t.Info.Iface.Methods {
				d[Id(endpointStructName(fn.Name))] = Qual(PackagePathGoKitTransportHTTP, "NewClient").Call(
					Line().Lit("POST"), // TODO: customize POST
					Line().Id("u"),
					Line().Qual(pathToHttpConverter(t.Info.ServiceImportPath), httpEncodeRequestName(fn)),
					Line().Qual(pathToHttpConverter(t.Info.ServiceImportPath), httpDecodeResponseName(fn)),
					Line().Id("opts").Op("...").Line(),
				).Dot("Endpoint").Call()
			}
		},
	)), Nil())
}
