package plugins

import (
	"bytes"
	"fmt"
	"path"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"unicode"

	. "github.com/dave/jennifer/jen"
	"github.com/devimteam/microgen/gen"
	mstrings "github.com/devimteam/microgen/gen/strings"
	"github.com/devimteam/microgen/internal"
	"github.com/devimteam/microgen/pkg/microgen"
	"github.com/devimteam/microgen/pkg/plugins/pkg"
	toml "github.com/pelletier/go-toml"
	"github.com/pkg/errors"
)

const (
	grpcKitPlugin = "go-kit-grpc"
)

type grpcGokitPlugin struct{}

type grpcGokitConfig struct {
	Path         string
	Protobuf     string
	TransportPkg string
	Client       struct {
		DefaultAddr string
	}
	CheckNil bool
}

func (p *grpcGokitPlugin) Generate(ctx microgen.Context, args []byte) (microgen.Context, error) {
	cfg := grpcGokitConfig{}
	if len(args) > 0 {
		err := toml.Unmarshal(args, &cfg)
		if err != nil {
			return ctx, err
		}
	}
	if cfg.Protobuf == "" {
		return ctx, errors.New("argument 'protobuf' is required")
	}
	if cfg.TransportPkg == "" {
		cfg.TransportPkg = "transport"
	}
	if cfg.Path == "" {
		cfg.Path = "transport/grpc"
	}
	resolvedPkgPath, err := gen.GetPkgPath(cfg.TransportPkg, true)
	if err != nil {
		return ctx, err
	}
	cfg.TransportPkg = resolvedPkgPath

	ctx, err = p.client(ctx, cfg)
	if err != nil {
		return ctx, err
	}
	ctx, err = p.server(ctx, cfg)
	if err != nil {
		return ctx, err
	}
	ctx, err = p.endpointConverters(ctx, cfg)
	if err != nil {
		return ctx, err
	}
	ctx, err = p.typeConverters(ctx, cfg)
	if err != nil {
		return ctx, err
	}
	return ctx, nil
}

func (p *grpcGokitPlugin) client(ctx microgen.Context, cfg grpcGokitConfig) (microgen.Context, error) {
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
	f.ImportAlias(cfg.Protobuf, "pb")
	f.ImportAlias(pkg.GoKitGRPC, "grpckit")
	f.HeaderComment(ctx.FileHeader)

	f.Func().Id("NewGRPCClient").
		ParamsFunc(func(p *Group) {
			p.Id("conn").Op("*").Qual(pkg.GoogleGRPC, "ClientConn")
			p.Id("addr").Id("string")
			p.Id("opts").Op("...").Qual(pkg.GoKitGRPC, "ClientOption")
		}).Qual(cfg.TransportPkg, _Endpoints_).
		BlockFunc(func(g *Group) {
			if cfg.Client.DefaultAddr != "" {
				g.If(Id("addr").Op("==").Lit("")).Block(
					Id("addr").Op("=").Lit(cfg.Client.DefaultAddr),
				)
			}
			g.Return().Qual(cfg.TransportPkg, _Endpoints_).Values(DictFunc(func(d Dict) {
				for _, fn := range ctx.Interface.Methods {
					if !ctx.AllowedMethods[fn.Name] {
						continue
					}
					client := &Statement{}
					client.Qual(pkg.GoKitGRPC, "NewClient").Call(
						Line().Id("conn"), Id("addr"), Lit(fn.Name),
						Line().Id(join_("_Encode", fn.Name, _Request_)),
						Line().Id(join_("_Decode", fn.Name, _Response_)),
						Line().Add(p.protobufResponseType(fn, cfg)).Values(),
						Line().Id("opts...").Line(),
					).Dot("Endpoint").Call()
					d[Id(join_(fn.Name, "Endpoint"))] = client
				}
			}))
		})
	f.Line().Func().Id("ClientOptionsBuilder").Params(
		Id("opts").Op("[]").Qual(pkg.GoKitGRPC, "ClientOption"),
		Id("fns...").Func().Params(Op("[]").Qual(pkg.GoKitGRPC, "ClientOption")).Params(Op("[]").Qual(pkg.GoKitGRPC, "ClientOption")),
	).Params(Op("[]").Qual(pkg.GoKitGRPC, "ClientOption")).Block(
		For().Id("i := range fns").Block(
			Id("opts = fns[i](opts)"),
		),
		Return(Id("opts")),
	)

	if ctx.Variables["trace"] == "true" {
		f.Line().Func().Id("TracingClientOptions").Params(
			Id("tracer").Qual(pkg.OpenTracing, "Tracer"),
			Id("logger").Qual(pkg.GoKitLog, "Logger"),
		).Params(
			Func().Params(Op("[]").Qual(pkg.GoKitGRPC, "ClientOption")).Params(Op("[]").Qual(pkg.GoKitGRPC, "ClientOption")),
		).Block(
			Return().Func().Params(Id("opts").Op("[]").Qual(pkg.GoKitGRPC, "ClientOption")).Params(Op("[]").Qual(pkg.GoKitGRPC, "ClientOption")).Block(
				Return().Append(Id("opts"), Qual(pkg.GoKitGRPC, "ClientBefore").Call(
					Line().Qual(pkg.GoKitOpenTracing, "ContextToGRPC").Call(Id("tracer"), Id("logger")).Op(",").Line(),
				)),
			),
		)
	}
	/*if cfg.Server.Trace {
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
	}*/

	outfile := microgen.File{
		Name: grpcKitPlugin,
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

func (p *grpcGokitPlugin) server(ctx microgen.Context, cfg grpcGokitConfig) (microgen.Context, error) {
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

	f.Type().Id(mstrings.ToLowerFirst(ctx.Interface.Name) + "Server").StructFunc(func(g *Group) {
		for _, method := range ctx.Interface.Methods {
			if !ctx.AllowedMethods[method.Name] {
				continue
			}
			g.Id(mstrings.ToLowerFirst(method.Name)).Qual(pkg.GoKitGRPC, "Handler")
		}
	}).Line()

	f.Func().Id("NewGRPCServer").
		ParamsFunc(func(p *Group) {
			p.Id("endpoints").Op("*").Qual(cfg.TransportPkg, _Endpoints_)
			if ctx.Variables["trace"] == "true" {
				p.Id("logger").Qual(pkg.GoKitLog, "Logger")
			}
			if ctx.Variables["trace"] == "true" {
				p.Id("tracer").Qual(pkg.OpenTracing, "Tracer")
			}
			p.Id("opts").Op("...").Qual(pkg.GoKitGRPC, "ServerOption")
		}).Params(
		Qual(cfg.Protobuf, mstrings.ToUpperFirst(ctx.Interface.Name)+"Server"),
	).
		Block(
			Return().Op("&").Id(mstrings.ToLowerFirst(ctx.Interface.Name) + "Server").Values(DictFunc(func(g Dict) {
				for _, m := range ctx.Interface.Methods {
					if !ctx.AllowedMethods[m.Name] {
						continue
					}
					g[(&Statement{}).Id(mstrings.ToLowerFirst(m.Name))] = Qual(pkg.GoKitGRPC, "NewServer").
						Call(
							Line().Id("endpoints").Dot(join_(m.Name, "Endpoint")),
							Line().Id(join_("_Decode", m.Name, _Request_)),
							Line().Id(join_("_Encode", m.Name, _Response_)),
							Line().Add(p.serverOpts(ctx, m)).Op("...").Line(),
						)
				}
			}),
			),
		)
	f.Line()

	for _, signature := range ctx.Interface.Methods {
		if !ctx.AllowedMethods[signature.Name] {
			continue
		}
		f.Add(p.grpcServerFunc(signature, ctx.Interface, cfg)).Line()
	}

	outfile := microgen.File{
		Name: grpcKitPlugin,
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

func (p *grpcGokitPlugin) endpointConverters(ctx microgen.Context, cfg grpcGokitConfig) (microgen.Context, error) {
	const filename = "endpoint_converters.microgen.go"
	ImportAliasFromSources = true
	pluginPackagePath, err := gen.GetPkgPath(filepath.Join(cfg.Path, filename), false)
	if err != nil {
		return ctx, errors.Wrap(err, filename)
	}
	pkgName, err := gen.PackageName(pluginPackagePath, "")
	if err != nil {
		return ctx, errors.Wrap(err, filename)
	}
	f := NewFilePathName(pluginPackagePath, pkgName)
	f.ImportAlias(ctx.SourcePackageImport, serviceAlias)
	f.ImportAlias(cfg.Protobuf, "pb")
	f.ImportAlias(pkg.GoKitGRPC, "grpckit")
	f.HeaderComment(ctx.FileHeader)

	var requestEncoders, responseEncoders, requestDecoders, responseDecoders []microgen.Method
	for _, fn := range ctx.Interface.Methods {
		if !ctx.AllowedMethods[fn.Name] {
			continue
		}
		requestEncoders = append(requestEncoders, fn)
		requestDecoders = append(requestDecoders, fn)
		responseEncoders = append(responseEncoders, fn)
		responseDecoders = append(responseDecoders, fn)
	}

	f.Comment("//========================================= Request Encoders =========================================//")
	for _, signature := range requestEncoders {
		if cfg.CheckNil {
			f.Line().Var().Id(endpointErrorName(signature, _Request_)).Op("=").Add(endpointError(signature, _Request_))
		}
		f.Line().Add(p.encodeRequest(ctx, signature, cfg))
	}
	f.Comment("//========================================= Request Decoders =========================================//")
	for _, signature := range requestDecoders {
		if cfg.CheckNil {
			f.Line().Var().Id(endpointErrorName(signature, _Request_)).Op("=").Add(endpointError(signature, _Request_))
		}
		f.Line().Add(p.decodeRequest(ctx, signature, cfg))
	}
	f.Comment("//========================================= Response Encoders =========================================//")
	for _, signature := range responseEncoders {
		if cfg.CheckNil {
			f.Line().Var().Id(endpointErrorName(signature, _Request_)).Op("=").Add(endpointError(signature, _Request_))
		}
		f.Line().Add(p.encodeResponse(ctx, signature, cfg))
	}
	f.Comment("//========================================= Response Decoders =========================================//")
	for _, signature := range responseDecoders {
		if cfg.CheckNil {
			f.Line().Var().Id(endpointErrorName(signature, _Request_)).Op("=").Add(endpointError(signature, _Request_))
		}
		f.Line().Add(p.decodeResponse(ctx, signature, cfg))
	}

	outfile := microgen.File{
		Name: grpcKitPlugin,
		Path: filepath.Join(cfg.Path, filename),
	}
	var b bytes.Buffer
	err = f.Render(&b)
	if err != nil {
		return ctx, errors.Wrap(err, filename)
	}
	outfile.Content = b.Bytes()
	ctx.Files = append(ctx.Files, outfile)
	return ctx, nil
}

func (p *grpcGokitPlugin) typeConverters(ctx microgen.Context, cfg grpcGokitConfig) (microgen.Context, error) {
	const filename = "type_converters.microgen.go"
	ImportAliasFromSources = true
	pluginPackagePath, err := gen.GetPkgPath(filepath.Join(cfg.Path, filename), false)
	if err != nil {
		return ctx, errors.Wrap(err, filename)
	}
	pkgName, err := gen.PackageName(pluginPackagePath, "")
	if err != nil {
		return ctx, errors.Wrap(err, filename)
	}
	f := NewFilePathName(pluginPackagePath, pkgName)
	f.ImportAlias(ctx.SourcePackageImport, serviceAlias)
	f.ImportAlias(cfg.Protobuf, "pb")
	f.ImportAlias(pkg.GoKitGRPC, "grpckit")
	f.HeaderComment(ctx.FileHeader)

	generated := make(map[reflect.Type][]string)
	for lenConvMap(generated) < lenConvMap(requiredConverters) {
		converters := listRequiredConverters()
		fmt.Println(lenConvMap(generated), "vs", len(converters))
		for _, c := range converters {
			if ss, ok := generated[c.t]; ok && mstrings.IsInStringSlice(c.name, ss) {
				continue
			}
			fmt.Println(c.t)
			f.Add(typeToProtoConverter(c.t, c.name, cfg))
			f.Add(typeFromProtoConverter(c.t, c.name, cfg))
			generated[c.t] = append(generated[c.t], c.name)
		}
	}

	outfile := microgen.File{
		Name: grpcKitPlugin,
		Path: filepath.Join(cfg.Path, filename),
	}
	var b bytes.Buffer
	err = f.Render(&b)
	if err != nil {
		return ctx, errors.Wrap(err, filename)
	}
	outfile.Content = b.Bytes()
	ctx.Files = append(ctx.Files, outfile)
	return ctx, nil
}

var ctx_contextContext = Id(_ctx_).Qual(pkg.Context, "Context")

func (p *grpcGokitPlugin) encodeRequest(ctx microgen.Context, fn microgen.Method, cfg grpcGokitConfig) *Statement {
	const fullName = "request"
	const shortName = "req"
	return genericEncode(_Response_, shortName, fullName, false)(fn, cfg)
}

func (p *grpcGokitPlugin) decodeRequest(ctx microgen.Context, fn microgen.Method, cfg grpcGokitConfig) *Statement {
	const fullName = "request"
	const shortName = "req"
	return genericDecode(_Request_, shortName, fullName, false)(fn, cfg)
}

func (p *grpcGokitPlugin) encodeResponse(ctx microgen.Context, fn microgen.Method, cfg grpcGokitConfig) *Statement {
	const fullName = "response"
	const shortName = "resp"
	return genericEncode(_Response_, shortName, fullName, true)(fn, cfg)
}

func (p *grpcGokitPlugin) decodeResponse(ctx microgen.Context, fn microgen.Method, cfg grpcGokitConfig) *Statement {
	const fullName = "response"
	const shortName = "resp"
	return genericDecode(_Response_, shortName, fullName, true)(fn, cfg)
}

func genericEncode(direction, shortName, fullName string, useResults bool) func(fn microgen.Method, cfg grpcGokitConfig) *Statement {
	return func(fn microgen.Method, cfg grpcGokitConfig) *Statement {
		methodParams := internal.RemoveContextIfFirst(fn.Args)
		if useResults {
			methodParams = internal.RemoveErrorIfLast(fn.Results)
		}
		return Line().Func().Id(join_("_Encode", fn.Name, direction)).Params(ctx_contextContext, Id(fullName).Interface()).
			Params(Interface(), Error()).BlockFunc(
			func(group *Group) {
				switch len(methodParams) {
				case 0:
					group.Return(Op("&").Qual(pkg.EmptyProtobuf, "Empty").Values(), Nil())
					return
				case 1:
					marshal, _, ok := findCustomBindingLayouts(methodParams[0].Type)
					if ok {
						group.Return(Id(fmt.Sprintf(marshal, fullName)), Nil())
						return
					}
					fallthrough
				default:
					if cfg.CheckNil {
						group.List(Id(shortName), Id("ok")).Op(":=").Id(fullName).Assert(Op("*").Qual(cfg.TransportPkg, join_(fn.Name, direction)))
						group.If(Id("!ok")).Block(
							Return(Nil(), Id(endpointErrorName(fn, direction))),
						)
					} else {
						group.Id(shortName).Op(":=").Id(fullName).Assert(Op("*").Qual(cfg.TransportPkg, join_(fn.Name, direction)))
					}
				}
				exchangeType := reflect.StructOf(makeStructFromVars(methodParams))
				group.Return().Id(converterToProtoName(exchangeType, join_(fn.Name, direction), true)).Call(Id(shortName))
			},
		).Line()
	}
}

func genericDecode(direction, shortName, fullName string, useResults bool) func(fn microgen.Method, cfg grpcGokitConfig) *Statement {
	return func(fn microgen.Method, cfg grpcGokitConfig) *Statement {
		methodParams := internal.RemoveContextIfFirst(fn.Args)
		if useResults {
			methodParams = internal.RemoveErrorIfLast(fn.Results)
		}
		return Line().Func().Id(join_("_Decode", fn.Name, direction)).Params(ctx_contextContext, Id(fullName).Interface()).
			Params(Interface(), Error()).BlockFunc(
			func(group *Group) {
				switch len(methodParams) {
				case 0:
					group.Return(Nil(), Nil())
					return
				case 1:
					_, unmarshal, ok := findCustomBindingLayouts(methodParams[0].Type)
					if ok {
						group.Return(Id(fmt.Sprintf(unmarshal, fullName)), Nil())
						return
					}
					fallthrough
				default:
					if cfg.CheckNil {
						group.List(Id(shortName), Id("ok")).Op(":=").Id(fullName).Assert(Op("*").Qual(cfg.Protobuf, join_(fn.Name, direction)))
						group.If(Id("!ok")).Block(
							Return(Nil(), Id(endpointErrorName(fn, direction))),
						)
					} else {
						group.Id(shortName).Op(":=").Id(fullName).Assert(Op("*").Qual(cfg.Protobuf, join_(fn.Name, direction)))
					}
				}
				exchangeType := reflect.StructOf(makeStructFromVars(methodParams))
				group.Return().Id(converterFromProtoName(exchangeType, join_(fn.Name, direction), true)).Call(Id(shortName))
			},
		).Line()
	}
}

func makeStructFromVars(vv []microgen.Var) []reflect.StructField {
	x := make([]reflect.StructField, len(vv))
	for i, v := range vv {
		x[i] = reflect.StructField{
			Name: mstrings.ToUpperFirst(v.Name),
			Type: v.Type,
		}
	}
	return x
}

var protobufCodec = reflect.TypeOf(new(ProtobufCodec)).Elem()

func typeToProtoConverter(t reflect.Type, structName string, cfg grpcGokitConfig) Code {
	s := &Statement{}
	currentType := func() func() *Statement {
		return func() *Statement {
			if structName != "" {
				return Qual(cfg.TransportPkg, structName)
			}
			return internal.VarType(t, false)
		}
	}()
	s.Func().Id(converterToProtoName(t, structName, true)).
		Params(ctx_contextContext, Id("value").Add(currentType())).
		Params(protobufType(t, structName, cfg), Error()).
		BlockFunc(func(body *Group) {
			switch t.Kind() {
			case reflect.Ptr:
				body.If(Id("value").Op("==").Nil()).Block(
					Return(Nil(), Nil()),
				).Line()
				body.Return(Id(converterToProtoName(t.Elem(), "", true)).Call(Op("*").Id("value")))
			case reflect.Slice:
				body.If(Id("value").Op("==").Nil()).Block(
					Return(Nil(), Nil()),
				)
				body.Var().Err().Error()
				body.Id("converted").Op(":=").Make(protobufType(t, structName, cfg), Len(Id("value")))
				body.For(Id(_i_).Op(":=").Range().Id("value")).BlockFunc(func(block *Group) {
					if t.Elem().Implements(protobufCodec) {
						block.List(Id("converted").Index(Id("i")), Err()).Op("=").Id("value").Index(Id(_i_)).Dot("ToProtobuf").Call()
					} else {
						block.List(Id("converted").Index(Id("i")), Err()).Op("=").Id(converterToProtoName(t.Elem(), "", true)).Call(Id("value").Index(Id(_i_)))
					}
					block.If(Err().Op("!=").Nil()).Block(
						Return(Nil(), Err()),
					)
				})
				body.Return(Id("converted"), Nil())
			case reflect.Struct:
				for fIdx, n := 0, t.NumField(); fIdx < n; fIdx++ {
					field := t.Field(fIdx)
					if field.Anonymous {
						continue
					}
					if r := []rune(field.Name)[0]; unicode.IsLower(r) {
						continue // unexported
					}
					if field.Type.Implements(protobufCodec) {
						body.List(Id("_"+field.Name), Err()).Op(":=").Id("value").Dot(field.Name).Dot("ToProtobuf").Call()
						body.If(Err().Op("!=").Nil()).Block(
							Return(Nil(), Err()),
						)
						continue
					}
					if fn, _, ok := findCustomBindingLayouts(field.Type); ok {
						body.List(Id("_" + field.Name)).Op(":=").Id(fmt.Sprintf(fn, "value."+field.Name))
						continue
					}
					body.List(Id("_"+field.Name), Err()).Op(":=").Id(converterToProtoName(field.Type, "", true)).Call(Id("value").Dot(field.Name))
					body.If(Err().Op("!=").Nil()).Block(
						Return(Nil(), Err()),
					)
				}
				body.Return(protobufType(t, structName, cfg).Values(DictFunc(func(d Dict) {
					for fIdx, n := 0, t.NumField(); fIdx < n; fIdx++ {
						field := t.Field(fIdx)
						if field.Anonymous {
							continue
						}
						if r := []rune(field.Name)[0]; unicode.IsLower(r) {
							continue // unexported
						}
						d[Id(field.Name)] = Id("_" + field.Name)
					}
				})), Nil())
			}
		})
	return s
}

func typeFromProtoConverter(t reflect.Type, structName string, cfg grpcGokitConfig) Code {
	s := &Statement{}
	currentType := func() func() *Statement {
		return func() *Statement {
			if structName != "" {
				return Qual(cfg.TransportPkg, structName)
			}
			return internal.VarType(t, false)
		}
	}()
	s.Func().Id(converterFromProtoName(t, structName, true)).
		Params(ctx_contextContext, Id("value").Add(protobufType(t, structName, cfg))).
		Params(currentType(), Error()).
		BlockFunc(func(body *Group) {
			switch t.Kind() {
			case reflect.Ptr:
				body.If(Id("value").Op("==").Nil()).Block(
					Return(Nil(), Nil()),
				).Line()
				body.Return(Id(converterFromProtoName(t.Elem(), "", true)).Call(Op("*").Id("value")))
			case reflect.Slice:
				body.If(Id("value").Op("==").Nil()).Block(
					Return(Nil(), Nil()),
				)
				body.Var().Err().Error()
				body.Id("converted").Op(":=").Make(currentType(), Len(Id("value")))
				body.For(Id(_i_).Op(":=").Range().Id("value")).BlockFunc(func(block *Group) {
					if t.Elem().Implements(protobufCodec) {
						block.List(Id("converted").Index(Id("i")), Err()).Op("=").Id("converted").Index(Id("i")).Dot("FromProtobuf").Call(Id("value").Index(Id(_i_)))
					} else {
						block.List(Id("converted").Index(Id("i")), Err()).Op("=").Id(converterFromProtoName(t.Elem(), "", true)).Call(Id("value").Index(Id(_i_)))
					}
					block.If(Err().Op("!=").Nil()).Block(
						Return(Nil(), Err()),
					)
				})
				body.Return(Id("converted"), Nil())
			case reflect.Struct:
				for fIdx, n := 0, t.NumField(); fIdx < n; fIdx++ {
					field := t.Field(fIdx)
					if field.Anonymous {
						continue
					}
					if r := []rune(field.Name)[0]; unicode.IsLower(r) {
						continue // unexported
					}
					if field.Type.Implements(protobufCodec) {
						body.List(Id("_"+field.Name), Err()).Op(":=").Id("value").Dot(field.Name).Dot("ToProtobuf").Call()
						body.If(Err().Op("!=").Nil()).Block(
							Return(Nil(), Err()),
						)
						continue
					}
					if fn, _, ok := findCustomBindingLayouts(field.Type); ok {
						body.List(Id("_" + field.Name)).Op(":=").Id(fmt.Sprintf(fn, "value."+field.Name))
						continue
					}
					body.List(Id("_"+field.Name), Err()).Op(":=").Id(converterFromProtoName(field.Type, "", true)).Call(Id("value").Dot(field.Name))
					body.If(Err().Op("!=").Nil()).Block(
						Return(Nil(), Err()),
					)
				}
				body.Return(currentType().Values(DictFunc(func(d Dict) {
					for fIdx, n := 0, t.NumField(); fIdx < n; fIdx++ {
						field := t.Field(fIdx)
						if field.Anonymous {
							continue
						}
						if r := []rune(field.Name)[0]; unicode.IsLower(r) {
							continue // unexported
						}
						d[Id(field.Name)] = Id("_" + field.Name)
					}
				})), Nil())
			}
		})
	return s
}

func protobufType(t reflect.Type, structName string, cfg grpcGokitConfig) *Statement {
	c := &Statement{}
Loop:
	for {
		if t.PkgPath() != "" {
			if structName != "" {
				c.Qual(cfg.Protobuf, structName)
			} else {
				c.Qual(cfg.Protobuf, t.Name())
			}
			break Loop
		}
		switch t.Kind() {
		case reflect.Array:
			c.Index(Lit(t.Len()))
			t = t.Elem()
		case reflect.Func:
			break Loop
		case reflect.Interface:
			if t.NumMethod() == 0 {
				c.Interface()
			} else if t == internal.ErrorType {
				c.Error()
			}
			break Loop
		case reflect.Ptr:
			c.Op("*")
			t = t.Elem()
		case reflect.Slice:
			c.Index()
			t = t.Elem()
		case reflect.Struct:
			c.Qual(cfg.Protobuf, structName)
			break Loop
		default:
			panic(errors.Errorf("unexpected type '%s' in 'protobufType' of kind '%s'", t.String(), t.Kind()))
		}
	}
	return c
}

func listRequiredConverters() requiredConvertersSlice {
	x := make(requiredConvertersSlice, 0)
	for directType, typeNames := range requiredConverters {
		for i := range typeNames {
			x = append(x, requiredConvertersType{name: typeNames[i], t: directType})
		}
	}
	sort.Sort(x)
	return x
}

func lenConvMap(someMap map[reflect.Type][]string) int {
	sum := 0
	for _, typeNames := range someMap {
		sum += len(typeNames)
	}
	return sum
}

type (
	requiredConvertersType struct {
		name string
		t    reflect.Type
	}
	requiredConvertersSlice []requiredConvertersType
)

func (s requiredConvertersSlice) Len() int {
	return len(s)
}

func (s requiredConvertersSlice) Less(i, j int) bool {
	return converterToProtoName(s[i].t, s[i].name, false) < converterToProtoName(s[j].t, s[j].name, false)
}

func (s requiredConvertersSlice) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

var requiredConverters = make(map[reflect.Type][]string)

func converterToProtoName(t reflect.Type, typeName string, register bool) string {
	if register {
		// check, if this type already registered
		if old, ok := requiredConverters[t]; ok {
			// check, if this name of anonymous type is already registered.
			if !mstrings.IsInStringSlice(typeName, old) {
				// append to existed list of anonymous types
				requiredConverters[t] = append(old, typeName)
			}
		} else {
			// not registered, create new
			requiredConverters[t] = []string{typeName}
		}
		fmt.Println("register", t, typeName, len(requiredConverters))
	}
	return genConvert("_ToProtobuf")(t, typeName)
}

func converterFromProtoName(t reflect.Type, typeName string, register bool) string {
	if register {
		// check, if this type already registered
		if old, ok := requiredConverters[t]; ok {
			// check, if this name of anonymous type is already registered.
			if !mstrings.IsInStringSlice(typeName, old) {
				// append to existed list of anonymous types
				requiredConverters[t] = append(old, typeName)
			}
		} else {
			// not registered, create new
			requiredConverters[t] = []string{typeName}
		}
		fmt.Println("register", t, typeName, len(requiredConverters))
	}
	return genConvert("_FromProtobuf")(t, typeName)
}

func genConvert(suffix string) func(reflect.Type, string) string {
	return func(t reflect.Type, typeName string) string {
		str := strings.Builder{}
		str.WriteRune('_')
	Loop:
		for {
			if t.PkgPath() != "" {
				str.WriteRune('_')
				str.WriteString(path.Base(t.PkgPath()))
				str.WriteString(typeName)
				str.WriteString(mstrings.ToUpperFirst(t.Name()))
				break Loop
			}
			switch t.Kind() {
			case reflect.Slice:
				str.WriteRune('S')
				t = t.Elem()
			case reflect.Array:
				str.WriteRune('A')
				t = t.Elem()
			case reflect.Ptr:
				str.WriteRune('P')
				t = t.Elem()
			case reflect.Struct:
				str.WriteRune('_')
				str.WriteString(typeName)
				break Loop
			case reflect.String, reflect.Int:
				str.WriteString(t.Kind().String())
				break Loop
			default:
				panic(errors.Errorf("unexpected type '%s' in 'genConvert' of kind '%s'", t.String(), t.Kind()))
			}
		}
		str.WriteString(suffix)
		return str.String()
	}
}

func (p *grpcGokitPlugin) grpcServerFunc(signature microgen.Method, i *microgen.Interface, cfg grpcGokitConfig) *Statement {
	return Func().
		Params(Id(internal.Rec(mstrings.ToLowerFirst(i.Name)+"Server")).Op("*").Id(mstrings.ToLowerFirst(i.Name)+"Server")).
		Id(signature.Name).
		Call(ctx_contextContext, Id("req").Add(p.protobufRequestType(signature, cfg))).
		Params(p.protobufResponseType(signature, cfg), Error()).
		BlockFunc(p.grpcServerFuncBody(signature, i, cfg))
}

func (p *grpcGokitPlugin) protobufRequestType(fn microgen.Method, cfg grpcGokitConfig) *Statement {
	args := internal.RemoveContextIfFirst(fn.Args)
	if len(args) == 0 {
		return Op("*").Qual(pkg.EmptyProtobuf, "Empty")
	}
	if len(args) == 1 {
		binding, ok := findCustomBinding(args[0].Type)
		if ok {
			return internal.VarType(binding, false)
		}
	}
	return Op("*").Qual(cfg.Protobuf, fn.Name+_Request_)
}

func (p *grpcGokitPlugin) protobufResponseType(fn microgen.Method, cfg grpcGokitConfig) Code {
	results := internal.RemoveErrorIfLast(fn.Results)
	if len(results) == 0 {
		return Qual(pkg.EmptyProtobuf, "Empty")
	}
	if len(results) == 1 {
		binding, ok := findCustomBinding(results[0].Type)
		if ok {
			return internal.VarType(binding, false)
		}
	}
	return Qual(cfg.Protobuf, fn.Name+_Response_)
}

func (p *grpcGokitPlugin) grpcServerFuncBody(signature microgen.Method, i *microgen.Interface, cfg grpcGokitConfig) func(g *Group) {
	return func(g *Group) {
		g.List(Id("_"), Id("resp"), Err()).
			Op(":=").
			Id(internal.Rec(mstrings.ToLowerFirst(i.Name)+"Server")).Dot(mstrings.ToLowerFirst(signature.Name)).Dot("ServeGRPC").Call(Id("ctx"), Id("req"))

		g.If(Err().Op("!=").Nil()).Block(
			Return().List(Nil(), Err()),
		)

		g.Return().List(Id("resp").Assert(p.protobufResponseType(signature, cfg)), Nil())
	}
}

func (p *grpcGokitPlugin) serverOpts(ctx microgen.Context, fn microgen.Method) *Statement {
	s := &Statement{}
	if ctx.Variables["trace"] == "true" {
		s.Op("append(")
		defer s.Op(")")
	}
	s.Id("opts")
	if ctx.Variables["trace"] == "true" {
		s.Op(",").Qual(pkg.GoKitGRPC, "ServerBefore").Call(
			Line().Qual(pkg.GoKitOpenTracing, "GRPCToContext").Call(Id("tracer"), Lit(fn.Name), Id("logger")),
		)
	}
	return s
}

func endpointError(fn microgen.Method, entityType string) *Statement {
	return Qual(pkg.Errors, "New").Call(Lit("unexpected type of " + join_(fn.Name, entityType)))
}

func endpointErrorName(fn microgen.Method, entityType string) string {
	return join_(mstrings.ToLowerFirst(fn.Name), entityType)
}

type ProtobufTypeBinder func(origType reflect.Type) (pbType reflect.Type, marshalLayout, unmarshalLayout string, ok bool)

type protobufBinder struct {
	fn ProtobufTypeBinder
}

var protobufBindings = make([]protobufBinder, 0)

func init() {
	RegisterProtobufTypeBinding(stdBinding)
}

func RegisterProtobufTypeBinding(fn ProtobufTypeBinder) {
	protobufBindings = append(protobufBindings, protobufBinder{fn: fn})
}

func findCustomBinding(t reflect.Type) (reflect.Type, bool) {
	n := len(protobufBindings)
	for i := 0; i < n; i++ {
		if s, _, _, ok := protobufBindings[n-i-1].fn(t); ok {
			return s, true
		}
	}
	return nil, false
}

func findCustomBindingLayouts(t reflect.Type) (marshal string, unmarshal string, ok bool) {
	n := len(protobufBindings)
	for i := 0; i < n; i++ {
		if _, marshalLayout, unmarshalLayout, ok := protobufBindings[n-i-1].fn(t); ok {
			return marshalLayout, unmarshalLayout, true
		}
	}
	return "", "", false
}

// Generic interface to use custom encoders and decoders golang <-> protobuf
type ProtobufCodec interface {
	// If type implements ProtobufEncoder then microgen would use function
	// ToProtobuf() (y Y, err error) of this type as a converter from golang to protobuf
	ProtobufEncoder()
	// If type implements ProtobufDecoder then microgen would use function
	// FromProtobuf(y Y) (x X, err error) of this type as a converter from protobuf to golang
	ProtobufDecoder()
}

func stdBinding(t reflect.Type) (reflect.Type, string, string, bool) {
	fmt.Println("check", t, intType, t == intType)
	switch t {
	case stringType, bytesType, intType:
		return t, "%s", "%s", true
	default:
		return nil, "", "", false
	}
}

var (
	stringType = reflect.TypeOf(new(string)).Elem()
	bytesType  = reflect.TypeOf(new([]byte)).Elem()
	intType    = reflect.TypeOf(new(int)).Elem()
)
