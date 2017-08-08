package middleware

import (
	. "github.com/dave/jennifer/jen"
	. "github.com/devimteam/microgen/generator/template"
	"github.com/devimteam/microgen/parser"
	"github.com/devimteam/microgen/util"
)

type LoggingTemplate struct {
}

func logging(str string) string {
	return str + "Logging"
}

var (
	loggerFiled    = "logger"
	loggerArg      = "logger"
	nextField      = "next"
	nextArg        = "next"
	serviceLogging = "ServiceLogging"
)

func (LoggingTemplate) Render(i *parser.Interface) *File {
	f := NewFile(i.PackageName)

	f.Func().Id(serviceLogging).Params(Id(loggerArg).Qual(PackageAliasGoKitLog, "Logger")).Params(Id(MiddlewareTypeName)).
		Block(newLoggingBody(i))

	f.Line()

	// Render type logger
	f.Type().Id(logging(i.Name)).Struct(
		Id(loggerFiled).Qual(PackageAliasGoKitLog, "Logger"),
		Id(nextField).Qual(i.PackageName, i.Name),
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

// Render body for new logging middleware
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
		Id(nextArg).Qual(i.PackageName, i.Name),
	).Params(
		Qual(i.PackageName, i.Name),
	).BlockFunc(func(g *Group) {
		g.Return(Id(logging(i.Name)).Values(
			Dict{
				Id(loggerFiled): Id(loggerArg),
				Id(nextField):   Id(nextArg),
			},
		))
	}))
}

func loggingFunc(signature *parser.FuncSignature, i *parser.Interface) *Statement {
	return MethodDefinition(logging(i.Name), signature).
		BlockFunc(loggingFuncBody(signature))
}

func loggingFuncBody(signature *parser.FuncSignature) func(g *Group) {
	inputParamsDesc := "Service arguments"
	outputParamsDesc := "Service results"
	method, begin, took := "method", "begin", "took"
	return func(g *Group) {
		g.Defer().Func().Params(Id(begin).Qual(PackageAliasTime, "Time")).Block(
			Id(util.FirstLowerChar(serviceLogging)).Dot(loggerFiled).Dot("Log").Call(
				Lit(method), Lit(signature.Name),
				Line().Comment(inputParamsDesc).
					Line().Add(paramsNameAndValue(signature.Params)),
				Line().Comment(outputParamsDesc).
					Line().Add(paramsNameAndValue(signature.Results)),
				Line().Lit(took), Qual(PackageAliasTime, "Since").Call(Id(begin)),
			),
		).Call(Qual(PackageAliasTime, "Now").Call())
	}
}

// Renders key/value pairs wrapped in Dict for provided fields.
//
//		Err:    err,
//		Result: result,
//
func paramsNameAndValue(fields []*parser.FuncField) *Statement {
	return ListFunc(func(g *Group) {
		for i, field := range fields {
			if field.Package != nil && field.Package.Path == PackageAliasContext && i == 0 {
				continue
			}
			g.List(Lit(field.Name), Id(field.Name))
			//d[structFieldName(field)] = Id(util.ToLowerFirst(field.Name))
		}
	})
}
