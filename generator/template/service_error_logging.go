package template

import (
	"context"

	. "github.com/dave/jennifer/jen"
	mstrings "github.com/devimteam/microgen/generator/strings"
	"github.com/devimteam/microgen/generator/write_strategy"
	"github.com/vetcher/go-astra/types"
)

const (
	serviceErrorLoggingStructName = "errorLoggingMiddleware"
)

var ServiceErrorLoggingMiddlewareName = mstrings.ToUpperFirst(serviceErrorLoggingStructName)

type errorLoggingTemplate struct {
	info *GenerationInfo
}

func NewErrorLoggingTemplate(info *GenerationInfo) Template {
	return &errorLoggingTemplate{
		info: info,
	}
}

func (t *errorLoggingTemplate) Render(ctx context.Context) write_strategy.Renderer {
	f := NewFile("service")
	f.ImportAlias(t.info.SourcePackageImport, serviceAlias)
	f.HeaderComment(t.info.FileHeader)

	f.Comment("ErrorLoggingMiddleware writes to logger any error, if it is not nil.").
		Line().Func().Id(ServiceErrorLoggingMiddlewareName).Params(Id(_logger_).Qual(PackagePathGoKitLog, "Logger")).Params(Id(MiddlewareTypeName)).
		Block(t.newRecoverBody(t.info.Iface))

	f.Line()

	// Render type logger
	f.Type().Id(serviceErrorLoggingStructName).Struct(
		Id(_logger_).Qual(PackagePathGoKitLog, "Logger"),
		Id(_next_).Qual(t.info.SourcePackageImport, t.info.Iface.Name),
	)

	// Render functions
	for _, signature := range t.info.Iface.Methods {
		f.Line()
		f.Add(t.recoverFunc(ctx, signature)).Line()
	}

	return f
}

func (errorLoggingTemplate) DefaultPath() string {
	return filenameBuilder(PathService, "error_logging")
}

func (t *errorLoggingTemplate) Prepare(ctx context.Context) error {
	return nil
}

func (t *errorLoggingTemplate) ChooseStrategy(ctx context.Context) (write_strategy.Strategy, error) {
	return write_strategy.NewCreateFileStrategy(t.info.OutputFilePath, t.DefaultPath()), nil
}

func (t *errorLoggingTemplate) newRecoverBody(i *types.Interface) *Statement {
	return Return(Func().Params(
		Id(_next_).Qual(t.info.SourcePackageImport, i.Name),
	).Params(
		Qual(t.info.SourcePackageImport, i.Name),
	).BlockFunc(func(g *Group) {
		g.Return(Op("&").Id(serviceErrorLoggingStructName).Values(
			Dict{
				Id(_logger_): Id(_logger_),
				Id(_next_):   Id(_next_),
			},
		))
	}))
}

func (t *errorLoggingTemplate) recoverFunc(ctx context.Context, signature *types.Function) *Statement {
	return methodDefinition(ctx, serviceErrorLoggingStructName, signature).
		BlockFunc(t.recoverFuncBody(signature))
}

func (t *errorLoggingTemplate) recoverFuncBody(signature *types.Function) func(g *Group) {
	return func(g *Group) {
		if !t.info.AllowedMethods[signature.Name] {
			s := &Statement{}
			if len(signature.Results) > 0 {
				s.Return()
			}
			s.Id(rec(serviceErrorLoggingStructName)).Dot(_next_).Dot(signature.Name).Call(paramNames(signature.Args))
			g.Add(s)
			return
		}
		g.Defer().Func().Params().Block(
			If(Id(nameOfLastResultError(signature)).Op("!=").Nil()).Block(
				Id(rec(serviceErrorLoggingStructName)).Dot(_logger_).Dot("Log").Call(
					Lit("method"), Lit(signature.Name),
					Lit("message"), Id(nameOfLastResultError(signature)),
				),
			),
		).Call()

		g.Return().Id(rec(serviceErrorLoggingStructName)).Dot(_next_).Dot(signature.Name).Call(paramNames(signature.Args))
	}
}
