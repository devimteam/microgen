package plugins

import (
	"bytes"

	. "github.com/dave/jennifer/jen"
	"github.com/devimteam/microgen/internal"
	ms "github.com/devimteam/microgen/internal/strings"
	"github.com/devimteam/microgen/pkg/microgen"
	"github.com/devimteam/microgen/pkg/plugins/pkg"
	toml "github.com/pelletier/go-toml"
)

// Recovering plugin generates interface closure that recovers panics from deeper calls and
// returns panic message as error
//
// Parameters:
//      - path : relative path of generated file. Default: `./recovering.microgen.go`
//      - name : generated closure name. Default: `RecoveringMiddleware`
//      - stack : logs panic stacktrace. Default: `false`
//
const RecoveringPlugin = "recovering"

type recoveringMiddlewarePlugin struct{}

type recoveringConfig struct {
	Path  string
	Name  string
	Stack bool
}

func (p *recoveringMiddlewarePlugin) Generate(ctx microgen.Context, args []byte) (microgen.Context, error) {
	cfg := recoveringConfig{}
	err := toml.Unmarshal(args, &cfg)
	if err != nil {
		return ctx, err
	}
	if cfg.Name == "" {
		cfg.Name = "RecoveringMiddleware"
	}
	if cfg.Path == "" {
		cfg.Path = "recovering.microgen.go"
	}
	outfile := microgen.File{}

	ImportAliasFromSources = true
	pluginPackagePath, err := internal.GetPkgPath(cfg.Path, false)
	if err != nil {
		return ctx, err
	}
	pkgName, err := internal.PackageName(pluginPackagePath, "")
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

	for _, signature := range ctx.Interface.Methods {
		f.Add(p.recoverFunc(ctx, cfg, signature)).Line()
	}

	outfile.Name = RecoveringPlugin
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

func (p *recoveringMiddlewarePlugin) recoverFunc(ctx microgen.Context, cfg recoveringConfig, signature microgen.Method) *Statement {
	return internal.MethodDefinition(ms.ToLowerFirst(cfg.Name), signature).
		BlockFunc(p.recoverFuncBody(ctx, cfg, signature))
}

func (p *recoveringMiddlewarePlugin) recoverFuncBody(ctx microgen.Context, cfg recoveringConfig, signature microgen.Method) func(g *Group) {
	return func(g *Group) {
		if !ctx.AllowedMethods[signature.Name] {
			s := &Statement{}
			if len(signature.Results) > 0 {
				s.Return()
			}
			s.Id(internal.Rec(ms.ToLowerFirst(cfg.Name))).Dot(_next_).Dot(signature.Name).Call(internal.ParamNames(signature.Args))
			g.Add(s)
			return
		}
		g.Defer().Func().Params().Block(
			If(Id("r").Op(":=").Recover(), Id("r").Op("!=").Nil()).BlockFunc(func(b *Group) {
				if cfg.Stack {
					// Repeated code from net/http serve function.
					//    const size = 64 << 10
					//    stack := make([]byte, size)
					//    stack = stack[:runtime.Stack(stack, false)]
					b.Const().Id("size").Op("=").Lit(64).Op("<<").Lit(10)
					b.Id("stack").Op(":=").Make(Index().Byte(), Id("size"))
					b.Id("stack").Op("=").Id("stack").Index(Empty(), Qual(pkg.Runtime, "Stack").Call(Id("stack"), False()))
				}
				b.Id(internal.Rec(ms.ToLowerFirst(cfg.Name))).Dot(_logger_).Dot("Log").CallFunc(func(call *Group) {
					call.Line().Lit("method")
					call.Lit(signature.Name)
					call.Line().Lit("message")
					call.Id("r")
					if cfg.Stack {
						call.Line().Lit("stack")
						call.Id("string(stack)")
					}
				})
				b.Id(internal.NameOfLastResultError(signature)).Op("=").Qual(pkg.FMT, "Errorf").Call(Lit("%v"), Id("r"))
			}),
		).Call()
		g.Return().Id(internal.Rec(ms.ToLowerFirst(cfg.Name))).Dot(_next_).Dot(signature.Name).Call(internal.ParamNames(signature.Args))
	}
}
