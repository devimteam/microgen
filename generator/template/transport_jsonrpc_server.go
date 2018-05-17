package template

import (
	"context"

	. "github.com/dave/jennifer/jen"
	"github.com/devimteam/microgen/generator/write_strategy"
	"github.com/devimteam/microgen/util"
	"github.com/vetcher/go-astra/types"
)

const (
	prefixJSONRPCAnnotationTag = "json-rpc-prefix"
	suffixJSONRPCAnnotationTag = "json-rpc-suffix"
)

type jsonrpcServerTemplate struct {
	info     *GenerationInfo
	prefixes map[string]string
	suffixes map[string]string
	tracing  bool
}

func NewJSONRPCServerTemplate(info *GenerationInfo) Template {
	return &jsonrpcServerTemplate{
		info: info,
	}
}

func (t *jsonrpcServerTemplate) DefaultPath() string {
	return "./transport/jsonrpc/server.go"
}

func (t *jsonrpcServerTemplate) ChooseStrategy(ctx context.Context) (write_strategy.Strategy, error) {
	return write_strategy.NewCreateFileStrategy(t.info.OutputFilePath, t.DefaultPath()), nil
}

func (t *jsonrpcServerTemplate) Prepare(ctx context.Context) error {
	t.prefixes = make(map[string]string)
	t.suffixes = make(map[string]string)
	for _, fn := range t.info.Iface.Methods {
		if s := util.FetchTags(fn.Docs, TagMark+prefixJSONRPCAnnotationTag); len(s) > 0 {
			t.prefixes[fn.Name] = s[0]
		}
		if s := util.FetchTags(fn.Docs, TagMark+suffixJSONRPCAnnotationTag); len(s) > 0 {
			t.suffixes[fn.Name] = s[0]
		}
	}
	tags := util.FetchTags(t.info.Iface.Docs, TagMark+MicrogenMainTag)
	for _, tag := range tags {
		switch tag {
		case TracingMiddlewareTag:
			t.tracing = true
		}
	}
	return nil
}

func (t *jsonrpcServerTemplate) Render(ctx context.Context) write_strategy.Renderer {
	f := NewFile("transportjsonrpc")
	f.ImportAlias(t.info.SourcePackageImport, serviceAlias)
	f.HeaderComment(t.info.FileHeader)
	f.PackageComment(`DO NOT EDIT.`)

	f.Type().Id(privateServerStructName(t.info.Iface)).StructFunc(func(g *Group) {
		for _, method := range t.info.Iface.Methods {
			g.Id(util.ToLowerFirst(method.Name)).Qual(PackagePathHttp, "Handler")
		}
	}).Line()

	f.Func().Id("NewJSONRPCServer").ParamsFunc(func(p *Group) {
		p.Id("endpoints").Op("*").Qual(t.info.SourcePackageImport, "Endpoints")
		if t.tracing {
			p.Id("logger").Qual(PackagePathGoKitLog, "Logger")
		}
		if t.tracing {
			p.Id("tracer").Qual(PackagePathOpenTracingGo, "Tracer")
		}
		p.Id("opts").Op("...").Qual(PackagePathGoKitTransportJSONRPC, "ServerOption")
	}).Params(
		Qual(PackagePathHttp, "Handler"),
	).Block(
		Return().Op("&").Id(privateServerStructName(t.info.Iface)).Values(DictFunc(func(g Dict) {
			for _, m := range t.info.Iface.Methods {
				g[(&Statement{}).Id(util.ToLowerFirst(m.Name))] = Qual(PackagePathGoKitTransportJSONRPC, "NewServer").
					Call(
						Line().Qual(PackagePathGoKitTransportJSONRPC, "EndpointCodecMap").Values(Dict{
							Line().Lit(t.prefixes[m.Name] + m.Name + t.suffixes[m.Name]): Qual(PackagePathGoKitTransportJSONRPC, "EndpointCodec").Values(Dict{
								Id("Endpoint"): Id("endpoints").Dot(endpointStructName(m.Name)),
								Id("Decode"):   Id(decodeRequestName(m)),
								Id("Encode"):   Id(encodeResponseName(m)),
							}),
						}),
						Line().Add(t.serverOpts(m)).Op("...").Line(),
					)
			}
		}),
		),
	)

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
