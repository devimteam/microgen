package plugins

import (
	"bytes"
	"encoding/json"

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
	Path   string
	Name   string
	Inline bool `json:",omitempty"`
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
	outfile := microgen.File{}

	calcParamAmount := func(name string, params []types.Variable) int {
		ignore := t.ignoreParams[name]
		lenParams := t.lenParams[name]
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
	requestStructName := func(signature *types.Function) string {
		return cfg.Name + signature.Name + "Request"
	}

	ImportAliasFromSources = true
	f := NewFilePathName(ctx.SourcePackageImport, ctx.SourcePackageName)
	f.ImportAlias(ctx.SourcePackageImport, serviceAlias)
	f.HeaderComment(ctx.FileHeader)

	Line().Func().Id(ms.ToUpperFirst(cfg.Name)).
		Params(Id(_logger_).Qual(pkg.GoKitLog, "Logger")).
		Params(Func().Params(Id(ctx.Interface.Name)).Params(Id(ctx.Interface.Name))).
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
		_ = fn
		f.Line()
	}

	if !cfg.Inline {
		if len(ctx.Interface.Methods) > 0 {
			f.Type().Op("(")
		}
		for _, signature := range ctx.Interface.Methods {
			if params := internal.RemoveContextIfFirst(signature.Args); calcParamAmount(signature.Name, params) > 0 {
				f.Add(t.loggingEntity(ctx, "log"+requestStructName(signature), signature, params))
			}
			if params := internal.RemoveErrorIfLast(signature.Results); calcParamAmount(signature.Name, params) > 0 {
				f.Add(t.loggingEntity(ctx, "log"+responseStructName(signature), signature, params))
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
