package plugins

import (
	"bytes"
	"fmt"
	"path/filepath"
	"strings"

	. "github.com/dave/jennifer/jen"
	go_case "github.com/devimteam/go-case"
	"github.com/devimteam/microgen/gen"
	mstrings "github.com/devimteam/microgen/generator/strings"
	"github.com/devimteam/microgen/internal"
	"github.com/devimteam/microgen/pkg/microgen"
	"github.com/devimteam/microgen/pkg/plugins/pkg"
	toml "github.com/pelletier/go-toml"
)

const (
	transportKitPlugin = "go-kit-transport"

	_Endpoints_ = "Endpoints"
)

type transportGokitPlugin struct{}

type transportGokitConfig struct {
	Path      string
	Exchanges struct {
		// When true, comment '//easyjson:json' above types will be generated
		Easyjson bool
		Style    string
	}
	Endpoints struct {
		Chain   bool
		Latency bool
	}
}

func (p *transportGokitPlugin) Generate(ctx microgen.Context, args []byte) (microgen.Context, error) {
	cfg := transportGokitConfig{}
	if len(args) > 0 {
		err := toml.Unmarshal(args, &cfg)
		if err != nil {
			return ctx, err
		}
	}
	if cfg.Path == "" {
		cfg.Path = "transport"
	}

	ctx, err := p.exchanges(ctx, cfg)
	if err != nil {
		return ctx, err
	}
	ctx, err = p.endpoints(ctx, cfg)
	if err != nil {
		return ctx, err
	}
	ctx, err = p.client(ctx, cfg)
	if err != nil {
		return ctx, err
	}
	ctx, err = p.server(ctx, cfg)
	if err != nil {
		return ctx, err
	}
	return ctx, nil
}

func (p *transportGokitPlugin) endpoints(ctx microgen.Context, cfg transportGokitConfig) (microgen.Context, error) {
	const filename = "endpoints.microgen.go"
	ImportAliasFromSources = true
	pluginPackagePath, err := gen.GetPkgPath(filepath.Join(cfg.Path, filename), false)
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

	f.Var().Id("_").Qual(ctx.SourcePackageImport, ctx.Interface.Name).Op("=&").Id(_Endpoints_).Block()

	f.Type().Id(_Endpoints_).StructFunc(func(g *Group) {
		for _, signature := range ctx.Interface.Methods {
			if !ctx.AllowedMethods[signature.Name] {
				continue
			}
			g.Id(join_(signature.Name, "Endpoint")).Qual(pkg.GoKitEndpoint, "Endpoint")
		}
	}).Line()

	for _, fn := range ctx.Interface.Methods {
		f.Add(p.serviceEndpointMethod(ctx, cfg, fn))
	}

	if cfg.Endpoints.Chain {
		f.Id(fmt.Sprintf(`
func %[1]sChain(fns ...func(%[1]s) %[1]s) func(%[1]s) %[1]s {
	n := len(fns)
	return func(endpoints %[1]s) %[1]s {
		for i := 0; i < n; i++ {
			// reverse order
			endpoints = fns[n-i-1](endpoints)
		}
		return endpoints
	}
}`, _Endpoints_))
	}
	if cfg.Endpoints.Latency {

		f.Func().Id("Latency").Params(
			Id("dur").Qual(pkg.GoKitMetrics, "Histogram"),
		).Params(
			Func().Params(Id("endpoints").Id(_Endpoints_)).Params(Id(_Endpoints_)),
		).Block(
			Return().Func().Params(Id("endpoints").Id(_Endpoints_)).Params(Id(_Endpoints_)).
				BlockFunc(func(body *Group) {
					body.Return(Id(_Endpoints_).Values(DictFunc(func(d Dict) {
						for _, signature := range ctx.Interface.Methods {
							if ctx.AllowedMethods[signature.Name] {
								// Produce code:
								// CreateComment_Endpoint:   latency(dur, "CreateComment")(endpoints.CreateComment_Endpoint),
								d[Id(join_(signature.Name, "Endpoint"))] = Id("latency").Call(Id("dur"),
									Lit(signature.Name)).Call(Id("endpoints").Dot(join_(signature.Name, "Endpoint")))
							}
						}
					})))
				}),
		)

		//		func latency(dur metrics.Histogram, methodName string) endpoint.Middleware {
		//			dur := dur.With("method", methodName)
		//			return func(next endpoint.Endpoint) endpoint.Endpoint {
		//				return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		//					defer func(begin time.Time) {
		//						dur.With("success", strconv.FormatBool(err == nil)).Observe(time.Since(begin).Seconds())
		//					}(time.Now())
		//					return next(ctx, request)
		//				}
		//			}
		//		}
		f.Func().Id("latency").Params(
			Id("dur").Qual(pkg.GoKitMetrics, "Histogram"), Id("methodName string"),
		).Params(
			Qual(pkg.GoKitEndpoint, "Middleware"),
		).Block(
			Id("dur").Op(":=").Id("dur").Dot("With").Call(Lit("method"), Id("methodName")),
			Return().Func().Params(Id(_next_).Qual(pkg.GoKitEndpoint, "Endpoint")).Params(Qual(pkg.GoKitEndpoint, "Endpoint")).Block(
				Return().Func().Params(
					Id(_ctx_).Qual(pkg.Context, "Context"), Id("request interface{}"),
				).Params(
					Id("request interface{}"), Id("err error"),
				).Block(
					Defer().Func().Params(Id("begin").Qual(pkg.Time, "Time")).Block(
						Id("dur").Dot("With").Call(Lit("success"), Qual(pkg.Strconv, "FormatBool").Call(Err().Op("==").Nil())).
							Dot("Observe").Call(Qual(pkg.Time, "Since").Call(Id("begin")).Dot("Seconds").Call()),
					).Call(Qual(pkg.Time, "Now").Call()),
					Return().Id(_next_).Call(Id(_ctx_), Id("request")),
				),
			),
		)
	}

	outfile := microgen.File{
		Name: transportKitPlugin,
		Path: filepath.Join(cfg.Path, filename),
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

func (p *transportGokitPlugin) serviceEndpointMethod(ctx microgen.Context, cfg transportGokitConfig, fn microgen.Method) *Statement {
	normal := internal.NormalizeFunction(fn)
	return internal.MethodDefinition(_Endpoints_, normal.Method).
		BlockFunc(p.serviceEndpointMethodBody(ctx, cfg, fn, normal.Method))
}

func (p *transportGokitPlugin) exchanges(ctx microgen.Context, cfg transportGokitConfig) (microgen.Context, error) {
	const filename = "exchanges.microgen.go"
	ImportAliasFromSources = true
	pluginPackagePath, err := gen.GetPkgPath(filepath.Join(cfg.Path, filename), false)
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
	if cfg.Exchanges.Easyjson {
		f.Id("//go:generate easyjson -all " + filename).Line()
	}

	for _, fn := range ctx.Interface.Methods {
		if !ctx.AllowedMethods[fn.Name] {
			continue
		}
		if cfg.Exchanges.Easyjson {
			f.Id("//easyjson:json")
		}
		requestName := join_(fn.Name, _Request_)
		f.Add(exchange(ctx, cfg, requestName, internal.RemoveContextIfFirst(fn.Args))).Line()
		if cfg.Exchanges.Easyjson {
			f.Id("//easyjson:json")
		}
		responseName := join_(fn.Name, _Response_)
		f.Add(exchange(ctx, cfg, responseName, internal.RemoveErrorIfLast(fn.Results))).Line()
	}

	outfile := microgen.File{
		Name: transportKitPlugin,
		Path: filepath.Join(cfg.Path, filename),
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

func (p *transportGokitPlugin) serviceEndpointMethodBody(ctx microgen.Context, cfg transportGokitConfig, fn microgen.Method, normal microgen.Method) func(g *Group) {
	reqName := "request"
	respName := "response"
	return func(g *Group) {
		if !ctx.AllowedMethods[fn.Name] {
			g.Return()
			return
		}
		requestName := join_(fn.Name, _Request_)
		responseName := join_(fn.Name, _Response_)
		endpointsStructFieldName := join_(fn.Name, "Endpoint")
		g.Id(reqName).Op(":=").Id(requestName).Values(internal.DictByNormalVariables(internal.RemoveContextIfFirst(fn.Args), internal.RemoveContextIfFirst(normal.Args)))
		g.Add(endpointResponse(respName, normal)).Id(internal.Rec(_Endpoints_)).Dot(endpointsStructFieldName).Call(Id(internal.FirstArgName(normal)), Op("&").Id(reqName))
		/*g.If(Id(nameOfLastResultError(normal)).Op("!=").Nil().BlockFunc(func(ifg *Group) {
			if internal.Tags(ctx).HasAny(GrpcTag, GrpcClientTag, GrpcServerTag) {
				ifg.Add(checkGRPCError(normal))
			}
			ifg.Return()
		}))*/
		g.ReturnFunc(func(group *Group) {
			for _, field := range internal.RemoveErrorIfLast(fn.Results) {
				group.Id(respName).Assert(Op("*").Id(responseName)).Dot(mstrings.ToUpperFirst(field.Name))
			}
			group.Id(internal.NameOfLastResultError(normal))
		})
	}
}

func endpointResponse(respName string, fn microgen.Method) *Statement {
	if len(internal.RemoveErrorIfLast(fn.Results)) > 0 {
		return List(Id(respName), Id(internal.NameOfLastResultError(fn))).Op(":=")
	}
	return List(Id("_"), Id(internal.NameOfLastResultError(fn))).Op("=")
}

func (p *transportGokitPlugin) client(ctx microgen.Context, cfg transportGokitConfig) (microgen.Context, error) {
	const filename = "client.microgen.go"
	ImportAliasFromSources = true
	pluginPackagePath, err := gen.GetPkgPath(filepath.Join(cfg.Path, filename), false)
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

	if ctx.Variables["trace"] == "true" {
		f.Func().Id("TraceClient").Params(
			Id("tracer").Qual(pkg.OpenTracing, "Tracer"),
		).Params(
			Func().Params(Id("endpoints").Id(_Endpoints_)).Params(Id(_Endpoints_)),
		).Block(
			Return().Func().Params(Id("endpoints").Id(_Endpoints_)).Params(Id(_Endpoints_)).
				BlockFunc(func(body *Group) {
					body.Return(Id(_Endpoints_).Values(DictFunc(func(d Dict) {
						for _, signature := range ctx.Interface.Methods {
							if ctx.AllowedMethods[signature.Name] {
								// CreateComment_Endpoint:   latency(dur, "CreateComment")(endpoints.CreateComment_Endpoint),
								d[Id(join_(signature.Name, "Endpoint"))] = Qual(pkg.GoKitOpenTracing, "TraceClient").Call(Id("tracer"),
									Lit(signature.Name)).Call(Id("endpoints").Dot(join_(signature.Name, "Endpoint")))
							}
						}
					})))
				}),
		)
	}

	outfile := microgen.File{
		Name: transportKitPlugin,
		Path: filepath.Join(cfg.Path, filename),
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

func (p *transportGokitPlugin) server(ctx microgen.Context, cfg transportGokitConfig) (microgen.Context, error) {
	const filename = "server.microgen.go"
	ImportAliasFromSources = true
	pluginPackagePath, err := gen.GetPkgPath(filepath.Join(cfg.Path, filename), false)
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

	if ctx.Variables["trace"] == "true" {
		f.Func().Id("TraceServer").Params(
			Id("tracer").Qual(pkg.OpenTracing, "Tracer"),
		).Params(
			Func().Params(Id("endpoints").Id(_Endpoints_)).Params(Id(_Endpoints_)),
		).Block(
			Return().Func().Params(Id("endpoints").Id(_Endpoints_)).Params(Id(_Endpoints_)).
				BlockFunc(func(body *Group) {
					body.Return(Id(_Endpoints_).Values(DictFunc(func(d Dict) {
						for _, signature := range ctx.Interface.Methods {
							if ctx.AllowedMethods[signature.Name] {
								// CreateComment_Endpoint:   latency(dur, "CreateComment")(endpoints.CreateComment_Endpoint),
								d[Id(join_(signature.Name, "Endpoint"))] = Qual(pkg.GoKitOpenTracing, "TraceServer").Call(Id("tracer"),
									Lit(signature.Name)).Call(Id("endpoints").Dot(join_(signature.Name, "Endpoint")))
							}
						}
					})))
				}),
		)
	}

	outfile := microgen.File{
		Name: transportKitPlugin,
		Path: filepath.Join(cfg.Path, filename),
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

// Renders exchanges that represents requests and responses.
//
//  type CreateVisitRequest struct {
//  	Visit *entity.Visit `json:"visit"`
//  }
//
func exchange(ctx microgen.Context, cfg transportGokitConfig, name string, params []microgen.Var) Code {
	if len(params) == 0 {
		return Comment("Formal exchange type, please do not delete.").Line().
			Type().Id(name).Struct()
	}
	return Type().Id(name).StructFunc(func(g *Group) {
		for _, param := range params {
			g.Add(structField(ctx, cfg, param))
		}
	})
}

func structField(ctx microgen.Context, cfg transportGokitConfig, field microgen.Var) *Statement {
	s := Id(mstrings.ToUpperFirst(field.Name))
	s.Add(internal.VarType(field.Type, false))
	s.Tag(map[string]string{"json": selectNamingFunc(cfg.Exchanges.Style)(field.Name)})
	/*if types.IsEllipsis(field.Type) {
		s.Comment("This field was defined with ellipsis (...).")
	}*/
	return s
}

func selectNamingFunc(name string) func(string) string {
	switch strings.ToLower(name) {
	case "golang", "go": // do nothing: write right how it was defined in interface
		return go_case.ToNoCase
	case "pascal": // CamelCase
		return go_case.ToCamelCase
	case "camel": // camelCase
		return go_case.ToCamelCaseLowerFirst
	case "allcap": // CAMEL_CASE
		return spipe(go_case.ToSnakeCase, strings.ToTitle)
	case "snake": // camel_case
		return mstrings.ToSnakeCase
	case "default":
		fallthrough
	default:
		return mstrings.ToUpperFirst
	}
}

func spipe(ff ...func(string) string) func(string) string {
	n := len(ff)
	return func(s string) string {
		for i := 0; i < n; i++ {
			s = ff[i](s)
		}
		return s
	}
}
