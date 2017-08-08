package middleware

import (
	. "github.com/dave/jennifer/jen"
	"github.com/devimteam/microgen/parser"
)

type LoggingTemplate struct {
}

func logging(str string) string {
	return str + "Logging"
}

var (
	loggerFiled = "logger"
	loggerArg   = "logger"
	nextField   = "next"
	nextArg     = "next"
)

func (LoggingTemplate) Render(i *parser.Interface) *File {
	f := NewFile(i.PackageName)

	f.Func().Id("ServiceLogging").Params(Id(loggerArg).Qual(PackageAliasGoKitLog, "Logger")).Params(Id(MiddlewareTypeName)).
		Block(newLoggingBody(i))

	f.Line()

	// Render type logger
	f.Type().Id(logging(i.Name)).Struct(
		Id(loggerFiled).Qual(PackageAliasGoKitLog, "Logger"),
		Id(nextField).Qual(i.PackageName, i.Name),
	)

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
