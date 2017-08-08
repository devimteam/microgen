package template

import (
	. "github.com/dave/jennifer/jen"
	"github.com/devimteam/microgen/parser"
	"github.com/devimteam/microgen/util"
)

type LoggingTemplate struct {
}

func loggingStructName(iface *parser.Interface) string {
	return iface.Name + "Logging"
}

const (
	loggerVar      = "logger"
	nextVar        = "next"
	serviceLogging = "ServiceLogging"
)

func (LoggingTemplate) Render(i *parser.Interface) *File {
	f := NewFile(i.PackageName)

	f.Func().Id(serviceLogging).Params(Id(loggerVar).Qual(PackagePathGoKitLog, "Logger")).Params(Id(MiddlewareTypeName)).
		Block(newLoggingBody(i))

	f.Line()

	// Render type logger
	f.Type().Id(loggingStructName(i)).Struct(
		Id(loggerVar).Qual(PackagePathGoKitLog, "Logger"),
		Id(nextVar).Qual(i.PackageName, i.Name),
	)

	// Render functions
	for _, signature := range i.FuncSignatures {
		f.Line()
		f.Add(loggingFunc(signature, i))
	}

	return f
}

func (LoggingTemplate) Path() string {
	return "./middleware/logging.go"
}

// Render body for new logging middleware.
//
//		return func(next stringsvc.StringService) stringsvc.StringService {
//			return StringServiceLogging{
//				logger: logger,
//				next:   next,
//			}
//		}
//
func newLoggingBody(i *parser.Interface) *Statement {
	return Return(Func().Params(
		Id(nextVar).Qual(i.PackageName, i.Name),
	).Params(
		Qual(i.PackageName, i.Name),
	).BlockFunc(func(g *Group) {
		g.Return(Id(loggingStructName(i)).Values(
			Dict{
				Id(loggerVar): Id(loggerVar),
				Id(nextVar):   Id(nextVar),
			},
		))
	}))
}

func loggingFunc(signature *parser.FuncSignature, i *parser.Interface) *Statement {
	return methodDefinition(loggingStructName(i), signature).
		BlockFunc(loggingFuncBody(signature))
}

func loggingFuncBody(signature *parser.FuncSignature) func(g *Group) {
	inputParamsDesc := "Service arguments"
	outputParamsDesc := "Service results"
	method, begin, took := "method", "begin", "took"
	return func(g *Group) {
		g.Defer().Func().Params(Id(begin).Qual(PackagePathTime, "Time")).Block(
			Id(util.FirstLowerChar(serviceLogging)).Dot(loggerVar).Dot("Log").Call(
				Lit(method), Lit(signature.Name),
				Line().Comment(inputParamsDesc).
					Line().Add(paramsNameAndValue(signature.Params)),
				Line().Comment(outputParamsDesc).
					Line().Add(paramsNameAndValue(signature.Results)),
				Line().Lit(took), Qual(PackagePathTime, "Since").Call(Id(begin)),
			),
		).Call(Qual(PackagePathTime, "Now").Call())
	}
}

// Renders key/value pairs wrapped in Dict for provided fields.
//
//		"err", err, "result", result,
//
func paramsNameAndValue(fields []*parser.FuncField) *Statement {
	return ListFunc(func(g *Group) {
		for i, field := range fields {
			if i == 0 && checkFieldIsContext(field) {
				continue
			}
			g.List(Lit(field.Name), Id(field.Name))
		}
	})
}
