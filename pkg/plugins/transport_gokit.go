package plugins

import (
	"bytes"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/devimteam/microgen/pkg/plugins/pkg"

	. "github.com/dave/jennifer/jen"
	go_case "github.com/devimteam/go-case"
	"github.com/devimteam/microgen/gen"
	mstrings "github.com/devimteam/microgen/generator/strings"
	"github.com/devimteam/microgen/internal"
	"github.com/devimteam/microgen/pkg/microgen"
	"github.com/vetcher/go-astra/types"
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

func (p *transportGokitPlugin) Generate(ctx microgen.Context, args json.RawMessage) (microgen.Context, error) {
	cfg := transportGokitConfig{}
	if len(args) > 0 {
		err := json.Unmarshal(args, &cfg)
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

	f.Type().Id(_Endpoints_).StructFunc(func(g *Group) {
		for _, signature := range ctx.Interface.Methods {
			if !ctx.AllowedMethods[signature.Name] {
				continue
			}
			g.Id(join_(signature.Name, "Endpoint")).Qual(pkg.GoKitEndpoint, "Endpoint")
		}
	}).Line()

	if cfg.Endpoints.Chain {
		//      func EndpointsChain(fns ...func(Endpoints) Endpoints) func(Endpoints) Endpoints {
		//      	n := len(fns)
		//      	return func(endpoints Endpoints) Endpoints {
		//      		for i := 0; i < n; i++ {
		//      			endpoints = fns[i](endpoints)
		//      		}
		//      		return endpoints
		//      	}
		//      }
		f.Id(fmt.Sprintf(`
func %[1]sChain(fns ...func(%[1]s) %[1]s) func(%[1]s) %[1]s {
	n := len(fns)
	return func(endpoints %[1]s) %[1]s {
		for i := 0; i < n; i++ {
			endpoints = fns[i](endpoints)
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
			Id("dur").Op(":=").Id("dur").Dot("With").Call(Id("method"), Id("methodName")),
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

// Renders exchanges that represents requests and responses.
//
//  type CreateVisitRequest struct {
//  	Visit *entity.Visit `json:"visit"`
//  }
//
func exchange(ctx microgen.Context, cfg transportGokitConfig, name string, params []types.Variable) Code {
	if len(params) == 0 {
		return Comment("Formal exchange type, please do not delete.").Line().
			Type().Id(name).Struct()
	}
	return Type().Id(name).StructFunc(func(g *Group) {
		for _, param := range params {
			g.Add(structField(ctx, cfg, &param))
		}
	})
}

func structField(ctx microgen.Context, cfg transportGokitConfig, field *types.Variable) *Statement {
	s := Id(mstrings.ToUpperFirst(field.Name))
	s.Add(internal.VarType(ctx, field.Type, false))
	s.Tag(map[string]string{"json": selectNamingFunc(cfg.Exchanges.Style)(field.Name)})
	if types.IsEllipsis(field.Type) {
		s.Comment("This field was defined with ellipsis (...).")
	}
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
