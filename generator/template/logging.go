package template

import (
	"github.com/devimteam/microgen/util"
	"github.com/vetcher/godecl/types"
	. "github.com/vetcher/jennifer/jen"
	"github.com/devimteam/microgen/generator/write_strategy"
)

const (
	loggerVarName            = "logger"
	nextVarName              = "next"
	serviceLoggingStructName = "serviceLogging"
)

type loggingTemplate struct {
	Info *GenerationInfo
}

func NewLoggingTemplate(info *GenerationInfo) Template {
	return &loggingTemplate{
		Info: info.Duplicate(),
	}
}

// Render all logging.go file.
//
//		// This file was automatically generated by "microgen" utility.
//		// Please, do not edit.
//		package middleware
//
//		import (
//			context "context"
//			svc "github.com/devimteam/microgen/test/svc"
//			log "github.com/go-kit/kit/log"
//			time "time"
//		)
//
//		func ServiceLogging(logger log.Logger) Middleware {
//			return func(next svc.StringService) svc.StringService {
//				return &serviceLogging{
//					logger: logger,
//					next:   next,
//				}
//			}
//		}
//
//		type serviceLogging struct {
//			logger log.Logger
//			next   svc.StringService
//		}
//
//		func (s *serviceLogging) Count(ctx context.Context, text string, symbol string) (count int, positions []int) {
//			defer func(begin time.Time) {
//				s.logger.Log(
//					"method", "Count",
//					"text", text,
// 					"symbol", symbol,
//					"count", count,
// 					"positions", positions,
//					"took", time.Since(begin))
//			}(time.Now())
//			return s.next.Count(ctx, text, symbol)
//		}
//
func (t *loggingTemplate) Render() write_strategy.Renderer {
	f := NewFile(t.Info.ServiceImportPackageName)
	f.PackageComment(FileHeader)
	f.PackageComment(`Please, do not edit.`)

	f.Func().Id(util.ToUpperFirst(serviceLoggingStructName)).Params(Id(loggerVarName).Qual(PackagePathGoKitLog, "Logger")).Params(Id(MiddlewareTypeName)).
		Block(t.newLoggingBody(t.Info.Iface))

	f.Line()

	// Render type logger
	f.Type().Id(serviceLoggingStructName).Struct(
		Id(loggerVarName).Qual(PackagePathGoKitLog, "Logger"),
		Id(nextVarName).Qual(t.Info.ServiceImportPath, t.Info.Iface.Name),
	)

	// Render functions
	for _, signature := range t.Info.Iface.Methods {
		f.Line()
		f.Add(loggingFunc(signature)).Line()
	}

	return f
}

func (loggingTemplate) DefaultPath() string {
	return "./middleware/logging.go"
}

func (t *loggingTemplate) ChooseStrategy() (write_strategy.Strategy, error) {
	return write_strategy.NewFileMethod(t.Info.AbsOutPath, t.DefaultPath()), nil
}

// Render body for new logging middleware.
//
//		return func(next svc.StringService) svc.StringService {
//			return &serviceLogging{
//				logger: logger,
//				next:   next,
//			}
//		}
//
func (t *loggingTemplate) newLoggingBody(i *types.Interface) *Statement {
	return Return(Func().Params(
		Id(nextVarName).Qual(t.Info.ServiceImportPath, i.Name),
	).Params(
		Qual(t.Info.ServiceImportPath, i.Name),
	).BlockFunc(func(g *Group) {
		g.Return(Op("&").Id(serviceLoggingStructName).Values(
			Dict{
				Id(loggerVarName): Id(loggerVarName),
				Id(nextVarName):   Id(nextVarName),
			},
		))
	}))
}

// Render logging middleware for interface method.
//
//		func (s *serviceLogging) Count(ctx context.Context, text string, symbol string) (count int, positions []int) {
//			defer func(begin time.Time) {
//				s.logger.Log(
//					"method", "Count",
//					"text", text, "symbol", symbol,
//					"count", count, "positions", positions,
//					"took", time.Since(begin))
//			}(time.Now())
//			return s.next.Count(ctx, text, symbol)
//		}
//
func loggingFunc(signature *types.Function) *Statement {
	return methodDefinition(serviceLoggingStructName, signature).
		BlockFunc(loggingFuncBody(signature))
}

// Render logging function body with request/response and time tracking.
//
//		defer func(begin time.Time) {
//			s.logger.Log(
//				"method", "Count",
//				"text", text, "symbol", symbol,
//				"count", count, "positions", positions,
//				"took", time.Since(begin))
//		}(time.Now())
//		return s.next.Count(ctx, text, symbol)
//
func loggingFuncBody(signature *types.Function) func(g *Group) {
	return func(g *Group) {
		g.Defer().Func().Params(Id("begin").Qual(PackagePathTime, "Time")).Block(
			Id(util.FirstLowerChar(serviceLoggingStructName)).Dot(loggerVarName).Dot("Log").Call(
				Line().Lit("method"), Lit(signature.Name),
				Add(paramsNameAndValue(removeContextIfFirst(signature.Args))),
				Add(paramsNameAndValue(removeContextIfFirst(signature.Results))),
				Line().Lit("took"), Qual(PackagePathTime, "Since").Call(Id("begin")),
			),
		).Call(Qual(PackagePathTime, "Now").Call())

		g.Return().Id(util.FirstLowerChar(serviceLoggingStructName)).Dot(nextVarName).Dot(signature.Name).Call(paramNames(signature.Args))
	}
}

// Renders key/value pairs wrapped in Dict for provided fields.
//
//		"err", err,
// 		"result", result,
//		"count", count,
//
func paramsNameAndValue(fields []types.Variable) *Statement {
	return ListFunc(func(g *Group) {
		for _, field := range fields {
			g.Line().List(Lit(field.Name), Id(field.Name))
		}
	})
}
