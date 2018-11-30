package plugins

import (
	"bytes"
	"path/filepath"
	"strings"

	"github.com/devimteam/microgen/internal/pkgpath"

	. "github.com/dave/jennifer/jen"
	"github.com/devimteam/microgen/internal"
	ms "github.com/devimteam/microgen/internal/strings"
	"github.com/devimteam/microgen/pkg/microgen"
	"github.com/devimteam/microgen/pkg/plugins/pkg"
	toml "github.com/pelletier/go-toml"
)

// Canonical plugin 'logging' generates interface closure that helps to log your method calls.
// It uses go-kit 'Logger' interface as parameter.
//
// Parameters:
//	- path : relative path of generated file. Default: `./logging.microgen.go`
//	- name : generated closure name. Default: `LoggingMiddleware`
//	- easyjson : add easyjson comment for each generated struct and general go:generate comment for automatization. Default: `false`
//	- took : additional field 'took' will be generated in response which will contain request duration. Default: `false`
//	- inline : generate all parameters (arguments and results) inlined in Log statement. Default: `false`
//
const LoggingPlugin = "logging"

type loggingMiddlewarePlugin struct{}

type loggingConfig struct {
	// Path to the desired file. By default './logging.go'.
	Path string
	// Name of middleware: structure, constructor and types prefixes.
	Name string
	// When true, all arguments and results will be on the same level,
	// instead of inside 'request' and 'response' fields. Also, special types will be omitted.
	Inline bool
	Ignore map[string][]string `toml:"-"`
	Len    map[string][]string `toml:"-"`
	// When true, comment '//easyjson:json' above types will be generated
	Easyjson bool
	// When true, additional field 'took' will be generated in response
	Took bool
}

func (p *loggingMiddlewarePlugin) Generate(ctx microgen.Context, args []byte) (microgen.Context, error) {
	cfg := loggingConfig{}
	err := toml.Unmarshal(args, &cfg)
	if err != nil {
		return ctx, err
	}
	if cfg.Name == "" {
		cfg.Name = "LoggingMiddleware"
	}
	if cfg.Path == "" {
		cfg.Path = "logging.microgen.go"
	}
	cfg.Ignore = makeMapFromComments("//logs-ignore", ctx.Interface)
	cfg.Len = makeMapFromComments("//logs-len", ctx.Interface)

	ImportAliasFromSources = true
	pluginPackagePath, err := pkgpath.GetPkgPath(cfg.Path, false)
	if err != nil {
		return ctx, err
	}
	pkgName, err := pkgpath.PackageName(pluginPackagePath, "")
	if err != nil {
		return ctx, err
	}
	f := NewFilePathName(pluginPackagePath, pkgName)
	f.ImportAlias(ctx.SourcePackageImport, serviceAlias)
	f.HeaderComment(ctx.FileHeader)

	filename := filepath.Base(cfg.Path)
	if cfg.Easyjson {
		f.Comment("//go:generate easyjson -all " + filename).Line()
	}

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
				f.Comment("//easyjson:json")
			}
			f.Type().Op("(")
		}
		for _, fn := range ctx.Interface.Methods {
			if params := internal.RemoveContextIfFirst(fn.Args); calcParamAmount(fn.Name, params, cfg) > 0 {
				f.Add(p.loggingEntity(ctx, join_("log", cfg.Name, fn.Name, _Request_), fn.Name, params, cfg))
			}
			if params := internal.RemoveErrorIfLast(fn.Results); calcParamAmount(fn.Name, params, cfg) > 0 {
				f.Add(p.loggingEntity(ctx, join_("log", cfg.Name, fn.Name, _Response_), fn.Name, params, cfg))
			}
		}
		if len(ctx.Interface.Methods) > 0 {
			f.Op(")")
		}
	}

	outfile := microgen.File{
		Name: LoggingPlugin,
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

func makeMapFromComments(prefix string, iface *microgen.Interface) map[string][]string {
	m := make(map[string][]string)
	for _, meth := range iface.Methods {
		m[meth.Name] = getListFromComments(prefix, meth.Docs)
	}
	return m
}

func getListFromComments(prefix string, comments []string) []string {
	var res []string
	for i := range comments {
		if !strings.HasPrefix(comments[i], prefix) {
			continue
		}
		res = append(res,
			strings.Split(
				strings.Replace(
					strings.Replace(
						comments[i][len(prefix):], " ", "", -1,
					),
					":", "", -1,
				),
				",",
			)...,
		)
	}
	return res
}

func (p *loggingMiddlewarePlugin) loggingEntity(ctx microgen.Context, name, fnName string, params []microgen.Var, cfg loggingConfig) Code {
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
				g.Id(ms.ToUpperFirst(field.Name)).Add(internal.VarType(field.Type, false))
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
func (p *loggingMiddlewarePlugin) loggingFunc(ctx microgen.Context, cfg loggingConfig, signature microgen.Method) *Statement {
	normal := internal.NormalizeFunction(signature)
	return internal.MethodDefinition(ms.ToLowerFirst(cfg.Name), normal.Method).
		BlockFunc(p.loggingFuncBody(ctx, cfg, signature))
}

func (p *loggingMiddlewarePlugin) loggingFuncBody(ctx microgen.Context, cfg loggingConfig, fn microgen.Method) func(g *Group) {
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
							Id(join_("log", cfg.Name, fn.Name, _Request_)).Add(
								p.fillMap(cfg,
									fn,
									internal.RemoveContextIfFirst(fn.Args),
									internal.RemoveContextIfFirst(normal.Args),
								),
							),
						)
					}
					if calcParamAmount(fn.Name, internal.RemoveErrorIfLast(fn.Results), cfg) > 0 {
						g.Line().List(
							Lit("response"),
							Id(join_("log", cfg.Name, fn.Name, _Response_)).Add(
								p.fillMap(cfg,
									fn,
									internal.RemoveErrorIfLast(fn.Results),
									internal.RemoveErrorIfLast(normal.Results),
								),
							),
						)
					}
				}
				if !ms.IsInStringSlice(internal.NameOfLastResultError(fn), cfg.Ignore[fn.Name]) {
					g.Line().List(Lit(internal.NameOfLastResultError(fn)), Id(internal.NameOfLastResultError(normal.Method)))
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
	return Id(join_("log", cfg.Name, fn.Name, suff)).Add(
		p.fillMap(cfg, fn.Parent, internal.RemoveContextIfFirst(fn.Parent.Args), internal.RemoveContextIfFirst(fn.Args)),
	)
}

func (p *loggingMiddlewarePlugin) fillMap(cfg loggingConfig, fn microgen.Method, params, normal []microgen.Var) *Statement {
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
func (p *loggingMiddlewarePlugin) paramsNameAndValue(cfg loggingConfig, fields, normFds []microgen.Var, functionName string) *Statement {
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

func calcParamAmount(name string, params []microgen.Var, cfg loggingConfig) int {
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

func join_(ss ...string) string {
	return strings.Join(ss, "_")
}
