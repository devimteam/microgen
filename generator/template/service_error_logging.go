package template

import (
	. "github.com/dave/jennifer/jen"
	"github.com/devimteam/microgen/generator/write_strategy"
	"github.com/devimteam/microgen/util"
	"github.com/vetcher/go-astra/types"
)

const (
	serviceErrorLoggingStructName = "serviceErrorLogging"
)

type errorLoggingTemplate struct {
	Info *GenerationInfo
}

func NewErrorLoggingTemplate(info *GenerationInfo) Template {
	return &errorLoggingTemplate{
		Info: info,
	}
}

func (t *errorLoggingTemplate) Render() write_strategy.Renderer {
	f := NewFile("middleware")
	f.ImportAlias(t.Info.SourcePackageImport, serviceAlias)
	f.PackageComment(t.Info.FileHeader)

	f.Comment("ServiceErrorLogging writes to logger any error, if it is not nil.").
		Line().Func().Id(util.ToUpperFirst(serviceErrorLoggingStructName)).Params(Id(loggerVarName).Qual(PackagePathGoKitLog, "Logger")).Params(Id(MiddlewareTypeName)).
		Block(t.newRecoverBody(t.Info.Iface))

	f.Line()

	// Render type logger
	f.Type().Id(serviceErrorLoggingStructName).Struct(
		Id(loggerVarName).Qual(PackagePathGoKitLog, "Logger"),
		Id(nextVarName).Qual(t.Info.SourcePackageImport, t.Info.Iface.Name),
	)

	// Render functions
	for _, signature := range t.Info.Iface.Methods {
		f.Line()
		f.Add(t.recoverFunc(signature)).Line()
	}

	return f
}

func (errorLoggingTemplate) DefaultPath() string {
	return filenameBuilder(PathService, "error_logging")
}

func (t *errorLoggingTemplate) Prepare() error {
	return nil
}

func (t *errorLoggingTemplate) ChooseStrategy() (write_strategy.Strategy, error) {
	return write_strategy.NewCreateFileStrategy(t.Info.AbsOutputFilePath, t.DefaultPath()), nil
}

func (t *errorLoggingTemplate) newRecoverBody(i *types.Interface) *Statement {
	return Return(Func().Params(
		Id(nextVarName).Qual(t.Info.SourcePackageImport, i.Name),
	).Params(
		Qual(t.Info.SourcePackageImport, i.Name),
	).BlockFunc(func(g *Group) {
		g.Return(Op("&").Id(serviceErrorLoggingStructName).Values(
			Dict{
				Id(loggerVarName): Id(loggerVarName),
				Id(nextVarName):   Id(nextVarName),
			},
		))
	}))
}

func (t *errorLoggingTemplate) recoverFunc(signature *types.Function) *Statement {
	return methodDefinition(serviceErrorLoggingStructName, signature).
		BlockFunc(t.recoverFuncBody(signature))
}

func (t *errorLoggingTemplate) recoverFuncBody(signature *types.Function) func(g *Group) {
	return func(g *Group) {
		g.Defer().Func().Params().Block(
			If(Id(nameOfLastResultError(signature)).Op("!=").Nil()).Block(
				Id(util.LastUpperOrFirst(serviceErrorLoggingStructName)).Dot(loggerVarName).Dot("Log").Call(
					Lit("method"), Lit(signature.Name),
					Lit("message"), Id(nameOfLastResultError(signature)),
				),
			),
		).Call()

		g.Return().Id(util.LastUpperOrFirst(serviceErrorLoggingStructName)).Dot(nextVarName).Dot(signature.Name).Call(paramNames(signature.Args))
	}
}
