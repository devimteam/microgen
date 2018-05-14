package template

import (
	"context"

	. "github.com/dave/jennifer/jen"
	"github.com/devimteam/microgen/generator/write_strategy"
	"github.com/devimteam/microgen/util"
	"github.com/vetcher/go-astra/types"
)

const (
	serviceRecoverStructName = "recoveringMiddleware"
)

type recoverTemplate struct {
	Info *GenerationInfo
}

func NewRecoverTemplate(info *GenerationInfo) Template {
	return &recoverTemplate{
		Info: info,
	}
}

func (t *recoverTemplate) Render(ctx context.Context) write_strategy.Renderer {
	f := NewFile("service")
	f.ImportAlias(t.Info.SourcePackageImport, serviceAlias)
	f.PackageComment(t.Info.FileHeader)

	f.Comment(util.ToUpperFirst(serviceRecoverStructName) + " recovers panics from method calls, writes to provided logger and returns the error of panic as method error.").
		Line().Func().Id(util.ToUpperFirst(serviceRecoverStructName)).Params(Id(loggerVarName).Qual(PackagePathGoKitLog, "Logger")).Params(Id(MiddlewareTypeName)).
		Block(t.newRecoverBody(t.Info.Iface))

	f.Line()

	// Render type logger
	f.Type().Id(serviceRecoverStructName).Struct(
		Id(loggerVarName).Qual(PackagePathGoKitLog, "Logger"),
		Id(nextVarName).Qual(t.Info.SourcePackageImport, t.Info.Iface.Name),
	)

	// Render functions
	for _, signature := range t.Info.Iface.Methods {
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
	return write_strategy.NewCreateFileStrategy(t.Info.AbsOutputFilePath, t.DefaultPath()), nil
}

func (t *recoverTemplate) newRecoverBody(i *types.Interface) *Statement {
	return Return(Func().Params(
		Id(nextVarName).Qual(t.Info.SourcePackageImport, i.Name),
	).Params(
		Qual(t.Info.SourcePackageImport, i.Name),
	).BlockFunc(func(g *Group) {
		g.Return(Op("&").Id(serviceRecoverStructName).Values(
			Dict{
				Id(loggerVarName): Id(loggerVarName),
				Id(nextVarName):   Id(nextVarName),
			},
		))
	}))
}

func (t *recoverTemplate) recoverFunc(ctx context.Context, signature *types.Function) *Statement {
	return methodDefinition(ctx, serviceRecoverStructName, signature).
		BlockFunc(t.recoverFuncBody(signature))
}

func (t *recoverTemplate) recoverFuncBody(signature *types.Function) func(g *Group) {
	return func(g *Group) {
		g.Defer().Func().Params().Block(
			If(Id("r").Op(":=").Recover(), Id("r").Op("!=").Nil()).Block(
				Id(util.LastUpperOrFirst(serviceRecoverStructName)).Dot(loggerVarName).Dot("Log").Call(
					Lit("method"), Lit(signature.Name),
					Lit("message"), Id("r"),
				),
				Id(nameOfLastResultError(signature)).Op("=").Qual(PackagePathFmt, "Errorf").Call(Lit("%v"), Id("r")),
			),
		).Call()

		g.Return().Id(util.LastUpperOrFirst(serviceRecoverStructName)).Dot(nextVarName).Dot(signature.Name).Call(paramNames(signature.Args))
	}
}
