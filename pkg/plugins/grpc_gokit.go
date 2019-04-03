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

	"github.com/gogo/protobuf/types"

	"github.com/cv21/microgen/internal"
	"github.com/cv21/microgen/internal/pkgpath"
	mstrings "github.com/cv21/microgen/internal/strings"
	"github.com/cv21/microgen/pkg/microgen"
	"github.com/cv21/microgen/pkg/plugins/pkg"
	. "github.com/dave/jennifer/jen"
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
	Proto    struct {
		Gen     bool
		Package string
		Path    string
	}
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
	resolvedPkgPath, err := pkgpath.GetPkgPath(cfg.TransportPkg, true)
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
	if cfg.Proto.Gen {
		ctx, err = p.protoFile(ctx, cfg)
		if err != nil {
			return ctx, err
		}
	}
	return ctx, nil
}

func (p *grpcGokitPlugin) client(ctx microgen.Context, cfg grpcGokitConfig) (microgen.Context, error) {
	const filename = "client.microgen.go"
	ImportAliasFromSources = true
	pluginPackagePath, err := pkgpath.GetPkgPath(filepath.Join(cfg.Path, filename), false)
	if err != nil {
		return ctx, err
	}
	pkgName, err := pkgpath.PackageName(pluginPackagePath, "")
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
						Line().New(p.protobufResponseType(fn, cfg)),
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
	pluginPackagePath, err := pkgpath.GetPkgPath(filepath.Join(cfg.Path, filename), false)
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
	pluginPackagePath, err := pkgpath.GetPkgPath(filepath.Join(cfg.Path, filename), false)
	if err != nil {
		return ctx, errors.Wrap(err, filename)
	}
	pkgName, err := pkgpath.PackageName(pluginPackagePath, "")
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
	pluginPackagePath, err := pkgpath.GetPkgPath(filepath.Join(cfg.Path, filename), false)
	if err != nil {
		return ctx, errors.Wrap(err, filename)
	}
	pkgName, err := pkgpath.PackageName(pluginPackagePath, "")
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
		//fmt.Println(lenConvMap(generated), "vs", len(converters))
		for _, c := range converters {
			if ss, ok := generated[c.t]; ok && mstrings.IsInStringSlice(c.name, ss) {
				continue
			}
			//fmt.Println(c.t)
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

func (p *grpcGokitPlugin) protoFile(ctx microgen.Context, cfg grpcGokitConfig) (microgen.Context, error) {
	const filename = "service.proto"
	const tab = "\t"
	f := internal.BufferAdapter{}
	if cfg.Proto.Path == "" {
		cfg.Proto.Path = filename
	}

	f.Ln(`syntax = "proto3";`)
	f.Ln()
	f.Lnf(`option go_package = "%s;pb";`, cfg.Protobuf)
	f.Ln()
	protoPkg := cfg.Proto.Package
	if protoPkg == "" {
		protoPkg = cfg.Client.DefaultAddr
		if strings.HasSuffix(protoPkg, ctx.Interface.Name) {
			protoPkg = protoPkg[:len(protoPkg)-len(ctx.Interface.Name)]
		}
	}

	f.Lnf("package %s;", protoPkg)
	f.Ln()
	{
		d := f.Hold()
		imports := make(map[string]struct{})
		// Draw service
		d.Lnf("service %s {", ctx.Interface.Name)
		for _, method := range ctx.Interface.Methods {
			if !ctx.AllowedMethods[method.Name] {
				continue
			}
			reqTypeName, externalImport := protoMessageName(internal.RemoveContextIfFirst(method.Args), join_(method.Name, _Request_))
			// we should replace builtin types because protobuf spec not allow constructions like 'rpc Method(string) returns (string)'
			reqTypeName, externalImport = replaceBuiltinProtoTypes(reqTypeName, externalImport)
			if externalImport != nil && *externalImport != "" {
				imports[*externalImport] = struct{}{}
			}
			respTypeName, externalImport := protoMessageName(internal.RemoveErrorIfLast(method.Results), join_(method.Name, _Response_))
			// we should replace builtin types because protobuf spec not allow constructions like 'rpc Method(string) returns (string)'
			respTypeName, externalImport = replaceBuiltinProtoTypes(respTypeName, externalImport)
			if externalImport != nil && *externalImport != "" {
				imports[*externalImport] = struct{}{}
			}
			d.Lnf(tab+"rpc %s (%s) returns (%s);", method.Name, reqTypeName, respTypeName)
		}
		d.Ln("}")

		// Draw message types
		for _, method := range ctx.Interface.Methods {
			if !ctx.AllowedMethods[method.Name] {
				continue
			}
			{
				args := internal.RemoveContextIfFirst(method.Args)
				reqTypeName, externalImport := protoMessageName(args, join_(method.Name, _Request_))
				if externalImport == nil && reqTypeName == join_(method.Name, _Request_) {
					d.Ln()
					d.Lnf("message %s {", reqTypeName)
					for i, arg := range args {
						n, imp := protoTypeName(arg.Type, "")
						if imp != nil {
							imports[*imp] = struct{}{}
						}
						d.Lnf(tab+"%s %s = %d;", n, mstrings.ToSnakeCase(arg.Name), i+1)
					}
					d.Ln("}")
				}
			}
			{
				params := internal.RemoveErrorIfLast(method.Results)
				reqTypeName, externalImport := protoMessageName(params, join_(method.Name, _Response_))
				if externalImport == nil && reqTypeName == join_(method.Name, _Response_) {
					d.Ln()
					d.Lnf("message %s {", reqTypeName)
					for i, arg := range params {
						n, imp := protoTypeName(arg.Type, "")
						if imp != nil {
							imports[*imp] = struct{}{}
						}
						d.Lnf(tab+"%s %s = %d;", n, mstrings.ToSnakeCase(arg.Name), i+1)
					}
					d.Ln("}")
				}
			}
		}
		{
			generated := make(map[reflect.Type][]string)
			for lenConvMap(generated) < lenConvMap(protobufMessages) {
				protobufMessages := listProtobufMessages()
				//fmt.Println(lenConvMap(generated), "vs", len(protobufMessages))
				for _, c := range protobufMessages {
					if ss, ok := generated[c.t]; ok && mstrings.IsInStringSlice(c.name, ss) {
						continue
					}
					//fmt.Println(c.t)
					d.Ln()
					d.Lnf("message %s {", c.name)
					for i, n := 0, c.t.NumField(); i < n; i++ {
						n, imp := protoTypeName(c.t.Field(i).Type, "")
						if imp != nil {
							imports[*imp] = struct{}{}
						}
						d.Lnf(tab+"%s %s = %d;", n, mstrings.ToSnakeCase(c.t.Field(i).Name), i+1)
					}
					d.Ln("}")
					generated[c.t] = append(generated[c.t], c.name)
				}
			}
		}

		for _, imp := range sortedSliceFromStringSet(imports) {
			f.Lnf(`import "%s";`, imp)
		}

		f.Ln()
		d.Release()
	}
	ctx.Files = append(ctx.Files, microgen.File{
		Path:    cfg.Proto.Path,
		Content: f.Bytes(),
	})
	return ctx, nil
}

var ctx_contextContext = Id(_ctx_).Qual(pkg.Context, "Context")

func (p *grpcGokitPlugin) encodeRequest(ctx microgen.Context, fn microgen.Method, cfg grpcGokitConfig) *Statement {
	const fullName = "request"
	const shortName = "req"
	return genericEncode(_Request_, shortName, fullName, false)(fn, cfg)
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
					if marshal, imp, ok := findCustomBindingMarshalLayout(methodParams[0].Type); ok {
						group.Return(qualOrId(imp, fmt.Sprintf(marshal, fullName)), Nil())
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
					if unmarshal, imp, ok := findCustomBindingUnmarshalLayout(methodParams[0].Type); ok {
						group.Return(qualOrId(imp, fmt.Sprintf(unmarshal, fullName)), Nil())
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
				)
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
					if fn, imp, ok := findCustomBindingMarshalLayout(field.Type); ok {
						body.List(Id("_" + field.Name)).Op(":=").Add(qualOrId(imp, fmt.Sprintf(fn, "value."+field.Name)))
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
				)
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
					if fn, imp, ok := findCustomBindingUnmarshalLayout(field.Type); ok {
						body.List(Id("_" + field.Name)).Op(":=").Add(qualOrId(imp, fmt.Sprintf(fn, "value."+field.Name)))
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
		if custom, found := findCustomBindingType(t); found {
			c.Add(internal.VarType(custom, false))
			break Loop
		}
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
		//fmt.Println("register", t, typeName, len(requiredConverters))
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
		//fmt.Println("register", t, typeName, len(requiredConverters))
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
			default:
				str.WriteString(t.Name())
				break Loop
				//panic(errors.Errorf("unexpected type '%s' in 'genConvert' of kind '%s'", t.String(), t.Kind()))
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
		Call(ctx_contextContext, Id("req").Add(Op("*").Add(p.protobufRequestType(signature, cfg)))).
		Params(Op("*").Add(p.protobufResponseType(signature, cfg)), Error()).
		BlockFunc(p.grpcServerFuncBody(signature, i, cfg))
}

func (p *grpcGokitPlugin) protobufRequestType(fn microgen.Method, cfg grpcGokitConfig) *Statement {
	args := internal.RemoveContextIfFirst(fn.Args)
	if len(args) == 0 {
		return Qual(pkg.EmptyProtobuf, "Empty")
	}
	if len(args) == 1 {
		binding, ok := findCustomBindingType(args[0].Type)
		if ok {
			return internal.VarType(binding, false)
		}
	}
	return Qual(cfg.Protobuf, fn.Name+_Request_)
}

func (p *grpcGokitPlugin) protobufResponseType(fn microgen.Method, cfg grpcGokitConfig) Code {
	results := internal.RemoveErrorIfLast(fn.Results)
	if len(results) == 0 {
		return Qual(pkg.EmptyProtobuf, "Empty")
	}
	if len(results) == 1 {
		binding, ok := findCustomBindingType(results[0].Type)
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

		g.Return().List(Id("resp").Assert(Op("*").Add(p.protobufResponseType(signature, cfg))), Nil())
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

const (
	googleProtobuf                  = "google.protobuf."
	googleProtobufEmpty             = googleProtobuf + "Empty"
	googleProtobufStringValueString = googleProtobuf + "StringValue"
	googleProtobufBoolValueString   = googleProtobuf + "BoolValue"
	googleProtobufInt64ValueString  = googleProtobuf + "Int64Value"

	importGoogleProtobuf         = "google/protobuf/"
	importGoogleProtobufWrappers = importGoogleProtobuf + "wrappers.proto"
	importGoogleProtobufEmpty    = importGoogleProtobuf + "empty.proto"
)

func protoMessageName(params []microgen.Var, def string) (string, *string) {
	switch len(params) {
	case 0:
		return googleProtobufEmpty, sp(importGoogleProtobufEmpty)
	case 1:
		if fieldType, requiredImport, ok := findCustomBindingProtoBinding(params[0].Type); ok {
			return fieldType, requiredImport
		}
		fallthrough
	default:
		return def, nil
	}
}

func replaceBuiltinProtoTypes(name string, imp *string) (string, *string) {
	if imp != nil {
		return name, imp
	}
	switch name {
	case "string":
		return googleProtobufStringValueString, sp(importGoogleProtobufWrappers)
	default:
		return name, imp
	}
}

var protobufMessages = make(map[reflect.Type][]string)

func listProtobufMessages() requiredConvertersSlice {
	x := make(requiredConvertersSlice, 0)
	for directType, typeNames := range protobufMessages {
		for i := range typeNames {
			x = append(x, requiredConvertersType{name: typeNames[i], t: directType})
		}
	}
	sort.Sort(x)
	return x
}

func protoTypeName(v reflect.Type, typeName string) (t string, imp *string) {
	if fieldType, requiredImport, ok := findCustomBindingProtoBinding(v); ok {
		return fieldType, requiredImport
	}
	if typeName == "" && v.Name() != "" {
		typeName = v.Name()
	}
	// check, if this type already registered
	if old, ok := protobufMessages[v]; ok {
		// check, if this name of anonymous type is already registered.
		if !mstrings.IsInStringSlice(typeName, old) {
			// append to existed list of anonymous types
			protobufMessages[v] = append(old, typeName)
		}
	} else {
		// not registered, create new
		if v.Kind() == reflect.Struct {
			protobufMessages[v] = []string{typeName}
		}
	}
	switch v {
	case stringType:
		return "string", nil
	case intType, int64Type:
		return "int64", nil
	case int32Type:
		return "int32", nil
	case uintType, uint64Type:
		return "uint64", nil
	case uint32Type:
		return "uint32", nil
	case bytesType:
		return "bytes", nil
	}
	switch v.Kind() {
	case reflect.Ptr:
		return protoTypeName(v.Elem(), typeName)
	case reflect.Slice:
		t, imp = protoTypeName(v.Elem(), typeName)
		return "repeated " + t, imp
	}
	return v.Name(), nil
}

func sp(s string) *string {
	return &s
}

func sortedSliceFromStringSet(s map[string]struct{}) []string {
	slice := make([]string, 0, len(s))
	for k := range s {
		slice = append(slice, k)
	}
	sort.Strings(slice)
	return slice
}

// todo: make better names
type ProtobufTypeBinder interface {
	ProtobufType(origType reflect.Type) (pbType reflect.Type, ok bool)
	MarshalLayout(origType reflect.Type) (marshalLayout string, requiredImport *string, ok bool)
	UnmarshalLayout(origType reflect.Type) (unmarshalLayout string, requiredImport *string, ok bool)
	ProtoBinding(origType reflect.Type) (fieldType string, requiredImport *string, ok bool)
}

var protobufBindings = make([]ProtobufTypeBinder, 0)

func init() {
	RegisterProtobufTypeBinding(stdBinding{})
}

func RegisterProtobufTypeBinding(fn ProtobufTypeBinder) {
	protobufBindings = append(protobufBindings, fn)
}

func findCustomBindingType(t reflect.Type) (reflect.Type, bool) {
	n := len(protobufBindings)
	for i := 0; i < n; i++ {
		if s, ok := protobufBindings[n-i-1].ProtobufType(t); ok {
			return s, true
		}
	}
	return nil, false
}

func findCustomBindingMarshalLayout(t reflect.Type) (marshal string, requiredImport *string, ok bool) {
	n := len(protobufBindings)
	for i := 0; i < n; i++ {
		if marshalLayout, imp, ok := protobufBindings[n-i-1].MarshalLayout(t); ok {
			return marshalLayout, imp, true
		}
	}
	return "", nil, false
}

func findCustomBindingUnmarshalLayout(t reflect.Type) (marshal string, requiredImport *string, ok bool) {
	n := len(protobufBindings)
	for i := 0; i < n; i++ {
		if unmarshalLayout, imp, ok := protobufBindings[n-i-1].UnmarshalLayout(t); ok {
			return unmarshalLayout, imp, true
		}
	}
	return "", nil, false
}

func findCustomBindingProtoBinding(t reflect.Type) (fieldType string, requiredImport *string, ok bool) {
	n := len(protobufBindings)
	for i := 0; i < n; i++ {
		if requiredType, requiredImport, ok := protobufBindings[n-i-1].ProtoBinding(t); ok {
			return requiredType, requiredImport, true
		}
	}
	return "", nil, false
}

type stdBinding struct{}

func (stdBinding) ProtobufType(origType reflect.Type) (pbType reflect.Type, ok bool) {
	switch origType {
	//case stringType:
	//	return googleProtobufStringValue, true
	case stringType, bytesType,
		intType, int32Type, int64Type,
		uintType, uint32Type, uint64Type,
		float32Type, float64Type, boolType:
		return origType, true
	default:
		return nil, false
	}
}

func (stdBinding) MarshalLayout(origType reflect.Type) (marshalLayout string, requiredImport *string, ok bool) {
	switch origType {
	case stringType, bytesType,
		intType, int32Type, int64Type,
		uintType, uint32Type, uint64Type,
		float32Type, float64Type, boolType:
		return "%s", nil, true
	default:
		return "", nil, false
	}
}

func (stdBinding) UnmarshalLayout(origType reflect.Type) (unmarshalLayout string, requiredImport *string, ok bool) {
	switch origType {
	case stringType, bytesType,
		intType, int32Type, int64Type,
		uintType, uint32Type, uint64Type,
		float32Type, float64Type, boolType:
		return "%s", nil, true
	default:
		return "", nil, false
	}
}

func (stdBinding) ProtoBinding(origType reflect.Type) (fieldType string, requiredImport *string, ok bool) {
	switch origType {
	case stringType, boolType,
		int32Type, int64Type,
		uint32Type, uint64Type:
		return origType.String(), nil, true
	case intType:
		return "int64", nil, true
	case uintType:
		return "uint64", nil, true
	case bytesType:
		return "bytes", nil, true
	case float32Type:
		return "float", nil, true
	case float64Type:
		return "double", nil, true
	default:
		return "", nil, false
	}
}

var (
	stringType  = reflect.TypeOf(new(string)).Elem()
	boolType    = reflect.TypeOf(new(bool)).Elem()
	bytesType   = reflect.TypeOf(new([]byte)).Elem()
	intType     = reflect.TypeOf(new(int)).Elem()
	int32Type   = reflect.TypeOf(new(int32)).Elem()
	int64Type   = reflect.TypeOf(new(int64)).Elem()
	uintType    = reflect.TypeOf(new(uint)).Elem()
	uint32Type  = reflect.TypeOf(new(uint32)).Elem()
	uint64Type  = reflect.TypeOf(new(uint64)).Elem()
	float32Type = reflect.TypeOf(new(float32)).Elem()
	float64Type = reflect.TypeOf(new(float64)).Elem()
)

var (
	googleProtobufStringValue = reflect.TypeOf(new(types.StringValue)).Elem()
)

// Generic interface to use custom encoders and decoders golang <-> protobuf
type ProtobufCodec interface {
	// If type implements ProtobufEncoder then microgen would use function
	// ToProtobuf() (y Y, err error) of this type as a converter from golang to protobuf
	ProtobufEncoder()
	// If type implements ProtobufDecoder then microgen would use function
	// FromProtobuf(y Y) (x X, err error) of this type as a converter from protobuf to golang
	ProtobufDecoder()
}

func qualOrId(imp *string, name string) Code {
	if imp != nil {
		return Qual(*imp, name)
	}
	return Id(name)
}
