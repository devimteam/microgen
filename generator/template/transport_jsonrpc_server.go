package template

import (
	. "github.com/dave/jennifer/jen"
	"github.com/devimteam/microgen/generator/write_strategy"
	"github.com/devimteam/microgen/util"
	"github.com/vetcher/godecl/types"
)

type jsonrpcServerTemplate struct {
	Info    *GenerationInfo
	methods map[string]string
	paths   map[string]string
	tracing bool
}

func NewJSONRPCServerTemplate(info *GenerationInfo) Template {
	return &jsonrpcServerTemplate{
		Info: info.Copy(),
	}
}

func (t *jsonrpcServerTemplate) DefaultPath() string {
	return "./transport/jsonrpc/server.go"
}

func (t *jsonrpcServerTemplate) ChooseStrategy() (write_strategy.Strategy, error) {
	if err := util.StatFile(t.Info.AbsOutPath, t.DefaultPath()); !t.Info.Force && err == nil {
		return nil, nil
	}
	return write_strategy.NewCreateFileStrategy(t.Info.AbsOutPath, t.DefaultPath()), nil
}

func (t *jsonrpcServerTemplate) Prepare() error {
	t.methods = make(map[string]string)
	t.paths = make(map[string]string)
	for _, fn := range t.Info.Iface.Methods {
		t.methods[fn.Name] = FetchHttpMethodTag(fn.Docs)
		t.paths[fn.Name] = buildMethodPath(fn)
	}
	tags := util.FetchTags(t.Info.Iface.Docs, TagMark+MicrogenMainTag)
	for _, tag := range tags {
		switch tag {
		case TracingTag:
			t.tracing = true
		}
	}
	return nil
}

func (t *jsonrpcServerTemplate) Render() write_strategy.Renderer {
	f := NewFile("transportjsonrpc")
	f.PackageComment(t.Info.FileHeader)
	f.PackageComment(`Please, do not edit.`)

	f.Func().Id("NewJSONRPCHandler").ParamsFunc(func(p *Group) {
		p.Id("endpoints").Op("*").Qual(t.Info.ServiceImportPath, "Endpoints")
		if t.tracing {
			p.Id("logger").Qual(PackagePathGoKitLog, "Logger")
		}
		if t.tracing {
			p.Id("tracer").Qual(PackagePathOpenTracingGo, "Tracer")
		}
		p.Id("opts").Op("...").Qual(PackagePathGoKitTransportJSONRPC, "ServerOption")
	}).Params(
		Qual(PackagePathHttp, "Handler"),
	).BlockFunc(func(g *Group) {
		g.Id("handler").Op(":=").Qual(PackagePathGoKitTransportJSONRPC, "NewServer").Call()
		for _, fn := range t.Info.Iface.Methods {
			g.Id("mux").Dot("Methods").Call(Lit(t.methods[fn.Name])).Dot("Path").
				Call(Lit("/" + t.paths[fn.Name])).Dot("Handler").Call(
				Line().Qual(PackagePathGoKitTransportJSONRPC, "NewServer").Call(
					Line().Id("endpoints").Dot(endpointStructName(fn.Name)),
					Line().Qual(pathToHttpConverter(t.Info.ServiceImportPath), decodeRequestName(fn)),
					Line().Qual(pathToHttpConverter(t.Info.ServiceImportPath), encodeResponseName(fn)),
					Line().Add(t.serverOpts(fn)).Op("...")),
			)
		}
		g.Return(Id("handler"))
	})

	return f
}

func (t *jsonrpcServerTemplate) serverOpts(fn *types.Function) *Statement {
	s := &Statement{}
	if t.tracing {
		s.Op("append(")
		defer s.Op(")")
	}
	s.Id("opts")
	if t.tracing {
		s.Op(",").Qual(PackagePathGoKitTransportJSONRPC, "ServerBefore").Call(
			Line().Qual(PackagePathGoKitTracing, "HTTPToContext").Call(Id("tracer"), Lit(fn.Name), Id("logger")),
		)
	}
	return s
}
