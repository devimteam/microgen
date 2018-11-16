package plugins

import (
	"bytes"

	"github.com/devimteam/microgen/internal"

	. "github.com/dave/jennifer/jen"
	"github.com/devimteam/microgen/gen"
	ms "github.com/devimteam/microgen/gen/strings"
	"github.com/devimteam/microgen/pkg/microgen"
	"github.com/devimteam/microgen/pkg/plugins/pkg"
	toml "github.com/pelletier/go-toml"
)

const opentracingPlugin = "opentracing"

type opentracingMiddlewarePlugin struct{}

type opentracingConfig struct {
	// Path to the desired file. By default './logging.go'.
	Path string
	// Name of middleware: structure, constructor and types prefixes.
	Name      string
	Type      TraceType
	Component string
}

type TraceType int

const (
	TraceServer TraceType = iota
	TraceClient
	TraceProducer
	TraceConsumer
)

func (p *opentracingMiddlewarePlugin) Generate(ctx microgen.Context, args []byte) (microgen.Context, error) {
	cfg := opentracingConfig{}
	err := toml.Unmarshal(args, &cfg)
	if err != nil {
		return ctx, err
	}
	if cfg.Name == "" {
		cfg.Name = "TracingMiddleware"
	}
	if cfg.Path == "" {
		cfg.Path = "tracing.microgen.go"
	}
	if cfg.Component == "" {
		if cfg.Type == TraceServer || cfg.Type == TraceConsumer {
			cfg.Component = ctx.Interface.Name
		}
	}

	ImportAliasFromSources = true
	pluginPackagePath, err := gen.GetPkgPath(cfg.Path, false)
	if err != nil {
		return ctx, err
	}
	pkgName, err := gen.PackageName(pluginPackagePath, "")
	if err != nil {
		return ctx, err
	}
	f := NewFilePathName(pluginPackagePath, pkgName)
	f.ImportAlias(ctx.SourcePackageImport, serviceAlias)
	f.HeaderComment(ctx.FileHeader)

	f.Var().Id("_").Qual(ctx.SourcePackageImport, ctx.Interface.Name).Op("=&").Id(ms.ToLowerFirst(cfg.Name)).Block()

	f.Line().Func().Id(ms.ToUpperFirst(cfg.Name)).
		Params(Id(_tracer_).Qual(pkg.OpenTracing, "Tracer")).
		Params(Func().Params(Qual(ctx.SourcePackageImport, ctx.Interface.Name)).Params(Qual(ctx.SourcePackageImport, ctx.Interface.Name))).
		Block(
			Return(Func().Params(
				Id(_next_).Qual(ctx.SourcePackageImport, ctx.Interface.Name),
			).Params(Qual(ctx.SourcePackageImport, ctx.Interface.Name)).
				Block(
					Return(Op("&").Id(ms.ToLowerFirst(cfg.Name)).Values(
						Dict{Id(_tracer_): Id(_tracer_), Id(_next_): Id(_next_)},
					)),
				),
			),
		)

	f.Line().Type().Id(ms.ToLowerFirst(cfg.Name)).Struct(
		Id(_tracer_).Qual(pkg.OpenTracing, "Tracer"),
		Id(_next_).Qual(ctx.SourcePackageImport, ctx.Interface.Name),
	)

	for _, fn := range ctx.Interface.Methods {
		f.Line().Add(p.tracingFunc(ctx, cfg, fn))
	}

	outfile := microgen.File{
		Name: opentracingPlugin,
		Path: cfg.Path,
	}
	var b bytes.Buffer
	err = f.Render(&b)
	if err != nil {
		return ctx, err
	}
	outfile.Content = b.Bytes()
	ctx.Files = append(ctx.Files, outfile)
	return ctx, nil
}

func (p *opentracingMiddlewarePlugin) tracingFunc(ctx microgen.Context, cfg opentracingConfig, fn microgen.Method) Code {
	normal := internal.NormalizeFunction(fn)
	return internal.MethodDefinition(ms.ToLowerFirst(cfg.Name), normal.Method).
		BlockFunc(p.tracingFuncBody(ctx, cfg, fn))
}

func (p *opentracingMiddlewarePlugin) tracingFuncBody(ctx microgen.Context, cfg opentracingConfig, fn microgen.Method) func(*Group) {
	normal := internal.NormalizeFunction(fn)
	rec := internal.Rec(ms.ToLowerFirst(cfg.Name))
	const _opSpan_, _parentSpan_ = "operationSpan", "parentSpan"
	return func(g *Group) {
		if !ctx.AllowedMethods[fn.Name] {
			s := &Statement{}
			if len(normal.Results) > 0 {
				s.Return()
			}
			s.Id(rec).Dot(_next_).Dot(fn.Name).Call(internal.ParamNames(normal.Args))
			g.Add(s)
			return
		}
		// todo: add special parameters
		var extendSpan, setTag, specialParameters Code
		switch cfg.Type {
		case TraceClient:
			extendSpan = Id(_opSpan_).Op("=").Id(rec).Dot(_tracer_).Dot("StartSpan").Call(Lit(fn.Name), Qual(pkg.OpenTracing, "ChildOf").Call(Id(_parentSpan_)))
			setTag = Qual(pkg.OpenTracingExt, "SpanKindRPCClient").Dot("Set").Call(Id(_opSpan_))
		case TraceServer:
			extendSpan = Id(_opSpan_).Op("=").Id(_parentSpan_).Dot("SetOperationName").Call(Lit(fn.Name))
			setTag = Qual(pkg.OpenTracingExt, "SpanKindRPCServer").Dot("Set").Call(Id(_opSpan_))
		case TraceProducer:
			extendSpan = Id(_opSpan_).Op("=").Id(rec).Dot(_tracer_).Dot("StartSpan").Call(Lit(fn.Name), Qual(pkg.OpenTracing, "FollowsFrom").Call(Id(_parentSpan_)))
			setTag = Qual(pkg.OpenTracingExt, "SpanKindProducer").Dot("Set").Call(Id(_opSpan_))
		case TraceConsumer:
			extendSpan = Id(_opSpan_).Op("=").Id(_parentSpan_).Dot("SetOperationName").Call(Lit(fn.Name))
			setTag = Qual(pkg.OpenTracingExt, "SpanKindConsumer").Dot("Set").Call(Id(_opSpan_))
		default:
			s := &Statement{}
			if len(normal.Results) > 0 {
				s.Return()
			}
			s.Id(rec).Dot(_next_).Dot(fn.Name).Call(internal.ParamNames(normal.Args))
			g.Add(s)
			return
		}
		g.Var().Id(_opSpan_).Qual(pkg.OpenTracing, "Span")
		g.If(Id(_parentSpan_).Op(":=").Qual(pkg.OpenTracing, "SpanFromContext").Call(Id(internal.FirstArgName(normal.Method))), Id(_parentSpan_).Op("!=").Nil()).Block(
			extendSpan,
		).Else().Block(
			Id(_opSpan_).Op("=").Id(rec).Dot(_tracer_).Dot("StartSpan").Call(Lit(fn.Name)),
		)
		g.Defer().Id(_opSpan_).Dot("Finish").Call()
		g.Add(setTag)
		if cfg.Component != "" {
			g.Qual(pkg.OpenTracingExt, "Component").Dot("Set").Call(Id(_opSpan_), Lit(cfg.Component))
		}
		g.Add(specialParameters)
		g.Return().Id(internal.Rec(ms.ToLowerFirst(cfg.Name))).Dot(_next_).Dot(fn.Name).Call(internal.ParamNames(normal.Args))
	}
}
