package template

import (
	"context"

	. "github.com/dave/jennifer/jen"
	mstrings "github.com/devimteam/microgen/generator/strings"
	"github.com/devimteam/microgen/generator/write_strategy"
	"github.com/vetcher/go-astra/types"
)

const (
	serviceRecoveringStructName = "recoveringMiddleware"
)

var ServiceRecoveringMiddlewareName = mstrings.ToUpperFirst(serviceRecoveringStructName)

type recoverTemplate struct {
	info *GenerationInfo
}

func NewRecoverTemplate(info *GenerationInfo) Template {
	return &recoverTemplate{
		info: info,
	}
}

func (t *recoverTemplate) Render(ctx context.Context) write_strategy.Renderer {
	f := NewFile("service")
	f.ImportAlias(t.info.SourcePackageImport, serviceAlias)
	f.HeaderComment(t.info.FileHeader)

	f.Comment(ServiceRecoveringMiddlewareName + " recovers panics from method calls, writes to provided logger and returns the error of panic as method error.").
		Line().Func().Id(ServiceRecoveringMiddlewareName).Params(Id(_logger_).Qual(PackagePathGoKitLog, "Logger")).Params(Id(MiddlewareTypeName)).
		Block(t.newRecoverBody(t.info.Iface))

	f.Line()

	// Render type logger
	f.Type().Id(serviceRecoveringStructName).Struct(
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

func (recoverTemplate) DefaultPath() string {
	return filenameBuilder(PathService, "recovering")
}

func (t *recoverTemplate) Prepare(ctx context.Context) error {
	return nil
}

func (t *recoverTemplate) ChooseStrategy(ctx context.Context) (write_strategy.Strategy, error) {
	return write_strategy.NewCreateFileStrategy(t.info.OutputFilePath, t.DefaultPath()), nil
}

func (t *recoverTemplate) newRecoverBody(i *types.Interface) *Statement {
	return Return(Func().Params(
		Id(_next_).Qual(t.info.SourcePackageImport, i.Name),
	).Params(
		Qual(t.info.SourcePackageImport, i.Name),
	).BlockFunc(func(g *Group) {
		g.Return(Op("&").Id(serviceRecoveringStructName).Values(
			Dict{
				Id(_logger_): Id(_logger_),
				Id(_next_):   Id(_next_),
			},
		))
	}))
}

func (t *recoverTemplate) recoverFunc(ctx context.Context, signature *types.Function) *Statement {
	return methodDefinition(ctx, serviceRecoveringStructName, signature).
		BlockFunc(t.recoverFuncBody(signature))
}

func (t *recoverTemplate) recoverFuncBody(signature *types.Function) func(g *Group) {
	return func(g *Group) {
		if !t.info.AllowedMethods[signature.Name] {
			s := &Statement{}
			if len(signature.Results) > 0 {
				s.Return()
			}
			s.Id(rec(serviceRecoveringStructName)).Dot(_next_).Dot(signature.Name).Call(paramNames(signature.Args))
			g.Add(s)
			return
		}
		g.Defer().Func().Params().Block(
			If(Id("r").Op(":=").Recover(), Id("r").Op("!=").Nil()).Block(
				Id(rec(serviceRecoveringStructName)).Dot(_logger_).Dot("Log").Call(
					Lit("method"), Lit(signature.Name),
					Lit("message"), Id("r"),
				),
				Id(nameOfLastResultError(signature)).Op("=").Qual(PackagePathFmt, "Errorf").Call(Lit("%v"), Id("r")),
			),
		).Call()
		g.Return().Id(rec(serviceRecoveringStructName)).Dot(_next_).Dot(signature.Name).Call(paramNames(signature.Args))
	}
}
