package plugins

import (
	"bytes"
	"encoding/json"
	"strings"

	"github.com/devimteam/microgen/gen"

	"github.com/vetcher/go-astra/types"

	"github.com/devimteam/microgen/internal"

	"github.com/devimteam/microgen/pkg/plugins/pkg"

	. "github.com/dave/jennifer/jen"
	ms "github.com/devimteam/microgen/generator/strings"
	"github.com/devimteam/microgen/pkg/microgen"
)

const loggingPlugin = "logging"

type loggingMiddlewarePlugin struct {
	Name string
}

type loggingConfig struct {
	Path     string
	Name     string
	Inline   bool
	Ignore   map[string][]string
	Len      map[string][]string
	Easyjson bool
	Took     bool
}

func (p *loggingMiddlewarePlugin) Generate(ctx microgen.Context, args json.RawMessage) (microgen.Context, error) {
	cfg := loggingConfig{}
	err := json.Unmarshal(args, &cfg)
	if err != nil {
		return ctx, err
	}
	if cfg.Name == "" {
		cfg.Name = "LoggingMiddleware"
	}
	if cfg.Path == "" {
		cfg.Path = "logging.go"
	}
	if cfg.Ignore == nil {
		cfg.Ignore = make(map[string][]string)
	}
	if cfg.Len == nil {
		cfg.Len = make(map[string][]string)
	}
	outfile := microgen.File{}

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
		Params(Id(_logger_).Qual(pkg.GoKitLog, "Logger")).
		Params(Func().Params(Qual(ctx.SourcePackageImport, ctx.Interface.Name)).Params(Qual(ctx.SourcePackageImport, ctx.Interface.Name))).
		Block(
			Return(Func().Params(
				Id(_next_).Qual(ctx.SourcePackageImport, ctx.Interface.Name),
			).Params(Qual(ctx.SourcePackageImport, ctx.Interface.Name)).
				Block(
					Return(Op("&").Id(ms.ToLowerFirst(cfg.Name)).Values(
						Dict{Id(_logger_): Id(_logger_), Id(_next_): Id(_next_)},
					)),
				),
			),
		)

	f.Line().Type().Id(ms.ToLowerFirst(cfg.Name)).Struct(
		Id(_logger_).Qual(pkg.GoKitLog, "Logger"),
		Id(_next_).Qual(ctx.SourcePackageImport, ctx.Interface.Name),
	)

	for _, fn := range ctx.Interface.Methods {
		f.Line().Add(p.loggingFunc(ctx, cfg, fn))
	}

	if !cfg.Inline {
		if len(ctx.Interface.Methods) > 0 {
			if cfg.Easyjson {
				f.Id("//easyjson:json")
			}
			f.Type().Op("(")
		}
		for _, fn := range ctx.Interface.Methods {
			if params := internal.RemoveContextIfFirst(fn.Args); calcParamAmount(fn.Name, params, cfg) > 0 {
				f.Add(p.loggingEntity(ctx, internalStructName("log", cfg.Name, fn.Name, _Request_), fn.Name, params, cfg))
			}
			if params := internal.RemoveErrorIfLast(fn.Results); calcParamAmount(fn.Name, params, cfg) > 0 {
				f.Add(p.loggingEntity(ctx, internalStructName("log", cfg.Name, fn.Name, _Response_), fn.Name, params, cfg))
			}
		}
		if len(ctx.Interface.Methods) > 0 {
			f.Op(")")
		}
	}

	outfile.Name = loggingPlugin
	outfile.Path = cfg.Path
	var b bytes.Buffer
	err = f.Render(&b)
	if err != nil {
		return ctx, err
	}
	outfile.Content = b.Bytes()
	ctx.Files = append(ctx.Files, outfile)
	return ctx, nil
}

func (p *loggingMiddlewarePlugin) loggingEntity(ctx microgen.Context, name, fnName string, params []types.Variable, cfg loggingConfig) Code {
	if len(params) == 0 {
		return Empty()
	}
	if !ctx.AllowedMethods[fnName] {
		return Empty()
	}
	s := &Statement{}
	return s.Id(name).StructFunc(func(g *Group) {
		ignore := cfg.Ignore[fnName]
		lenParams := cfg.Len[fnName]
		for _, field := range params {
			if !ms.IsInStringSlice(field.Name, ignore) {
				g.Id(ms.ToUpperFirst(field.Name)).Add(internal.VarType(ctx, field.Type, false))
			}
			if ms.IsInStringSlice(field.Name, lenParams) {
				g.Id("Len" + ms.ToUpperFirst(field.Name)).Int().Tag(map[string]string{"json": "len(" + ms.ToUpperFirst(field.Name) + ")"})
			}
		}
	})
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
func (p *loggingMiddlewarePlugin) loggingFunc(ctx microgen.Context, cfg loggingConfig, signature *types.Function) *Statement {
	normal := internal.NormalizeFunction(signature)
	return internal.MethodDefinition(ctx, ms.ToLowerFirst(cfg.Name), &normal.Function).
		BlockFunc(p.loggingFuncBody(ctx, cfg, signature))
}

func (p *loggingMiddlewarePlugin) loggingFuncBody(ctx microgen.Context, cfg loggingConfig, fn *types.Function) func(g *Group) {
	normal := internal.NormalizeFunction(fn)
	return func(g *Group) {
		if !ctx.AllowedMethods[fn.Name] {
			s := &Statement{}
			if len(normal.Results) > 0 {
				s.Return()
			}
			s.Id(internal.Rec(ms.ToLowerFirst(cfg.Name))).Dot(_next_).Dot(fn.Name).Call(internal.ParamNames(normal.Args))
			g.Add(s)
			return
		}
		tookFuncArgs, tookFuncCall := &Statement{}, &Statement{}
		if cfg.Took {
			tookFuncArgs = Id("begin").Qual(pkg.Time, "Time")
			tookFuncCall = Qual(pkg.Time, "Now").Call()
		}
		g.Defer().Func().Params(tookFuncArgs).Block(
			Id(internal.Rec(ms.ToLowerFirst(cfg.Name))).Dot(_logger_).Dot("Log").CallFunc(func(g *Group) {
				g.Line().Lit("method")
				g.Lit(fn.Name)
				if cfg.Inline {
					if calcParamAmount(fn.Name, internal.RemoveContextIfFirst(fn.Args), cfg) > 0 {
						g.Add(p.paramsNameAndValue(
							cfg, internal.RemoveContextIfFirst(fn.Args), internal.RemoveContextIfFirst(normal.Args), fn.Name),
						)
					}
					if calcParamAmount(fn.Name, internal.RemoveErrorIfLast(fn.Results), cfg) > 0 {
						g.Add(p.paramsNameAndValue(
							cfg, internal.RemoveErrorIfLast(fn.Results), internal.RemoveErrorIfLast(normal.Results), fn.Name),
						)
					}
				} else {
					if calcParamAmount(fn.Name, internal.RemoveContextIfFirst(fn.Args), cfg) > 0 {
						g.Line().List(
							Lit("request"),
							Id(internalStructName("log", cfg.Name, fn.Name, _Request_)).Add(
								p.fillMap(cfg,
									normal.Parent,
									internal.RemoveContextIfFirst(normal.Parent.Args),
									internal.RemoveContextIfFirst(fn.Args),
								),
							),
						) //p.loggerLogContent(cfg, normal, _Request_))
					}
					if calcParamAmount(fn.Name, internal.RemoveErrorIfLast(fn.Results), cfg) > 0 {
						g.Line().List(
							Lit("response"),
							Id(internalStructName("log", cfg.Name, fn.Name, _Response_)).Add(
								p.fillMap(cfg,
									normal.Parent,
									internal.RemoveErrorIfLast(normal.Parent.Results),
									internal.RemoveErrorIfLast(fn.Results),
								),
							),
						)
						//g.Line().List(Lit("response"), p.loggerLogContent(cfg, normal, _Response_))
					}
				}
				if !ms.IsInStringSlice(internal.NameOfLastResultError(fn), cfg.Ignore[fn.Name]) {
					g.Line().List(Lit(internal.NameOfLastResultError(fn)), Id(internal.NameOfLastResultError(&normal.Function)))
				}
				if cfg.Took {
					g.Line().List(Lit("took"), Qual(pkg.Time, "Since").Call(Id("begin")))
				}
			}),
		).Call(tookFuncCall)
		g.Return().Id(internal.Rec(ms.ToLowerFirst(cfg.Name))).Dot(_next_).Dot(fn.Name).Call(internal.ParamNames(normal.Args))
	}
}

func (p *loggingMiddlewarePlugin) loggerLogContent(
	cfg loggingConfig,
	fn *internal.NormalizedFunction,
	suff string,
) *Statement {
	return Id(internalStructName("log", cfg.Name, fn.Name, suff)).Add(
		p.fillMap(cfg, fn.Parent, internal.RemoveContextIfFirst(fn.Parent.Args), internal.RemoveContextIfFirst(fn.Args)),
	)
}

func (p *loggingMiddlewarePlugin) fillMap(cfg loggingConfig, fn *types.Function, params, normal []types.Variable) *Statement {
	return Values(DictFunc(func(d Dict) {
		ignore := cfg.Ignore[fn.Name]
		lenParams := cfg.Len[fn.Name]
		for i, field := range params {
			if !ms.IsInStringSlice(field.Name, ignore) {
				d[Id(ms.ToUpperFirst(field.Name))] = Id(normal[i].Name)
			}
			if ms.IsInStringSlice(field.Name, lenParams) {
				d[Id("Len"+ms.ToUpperFirst(field.Name))] = Len(Id(normal[i].Name))
			}
		}
	}))
}

// Renders key/value pairs wrapped in Dict for provided fields.
//
//		"err", err,
// 		"result", result,
//		"count", count,
//
func (p *loggingMiddlewarePlugin) paramsNameAndValue(cfg loggingConfig, fields, normFds []types.Variable, functionName string) *Statement {
	return ListFunc(func(g *Group) {
		ignore := cfg.Ignore[functionName]
		lenParams := cfg.Len[functionName]
		for i, field := range fields {
			if !ms.IsInStringSlice(field.Name, ignore) {
				g.Line().List(Lit(field.Name), Id(normFds[i].Name))
			}
			if ms.IsInStringSlice(field.Name, lenParams) {
				g.Line().List(Lit("len("+field.Name+")"), Len(Id(normFds[i].Name)))
			}
		}
	})
}

func calcParamAmount(name string, params []types.Variable, cfg loggingConfig) int {
	ignore := cfg.Ignore[name]
	lenParams := cfg.Len[name]
	paramAmount := len(params)
	for _, field := range params {
		if ms.IsInStringSlice(field.Name, ignore) {
			paramAmount -= 1
		}
		if ms.IsInStringSlice(field.Name, lenParams) {
			paramAmount += 1
		}
	}
	return paramAmount
}

func internalStructName(ss ...string) string {
	return strings.Join(ss, "_")
}

//
//func requestStructName(prefix string, name string, signature *types.Function, suffix string) string {
//	return fmt.Sprintf("_%s_%s_%s", name, signature.Name, "Request")
//}
//func responseStructName(name string, signature *types.Function) string {
//	return fmt.Sprintf("_%s_%s_%s", name, signature.Name, "Response")
//}
