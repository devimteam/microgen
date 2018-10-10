package plugins

import (
	"bytes"
	"encoding/json"
	"path/filepath"

	. "github.com/dave/jennifer/jen"
	"github.com/devimteam/microgen/gen"
	"github.com/devimteam/microgen/pkg/microgen"
)

const (
	grpcKitPlugin     = "go-kit-grpc"
	grpcKitPluginName = "grpc"
)

type grpcGokitPlugin struct{}

type grpcGokitConfig struct {
	Path      string
	Endpoints struct {
		// When true, comment '//easyjson:json' above types will be generated
		Easyjson bool
	}
}

func (p *grpcGokitPlugin) Generate(ctx microgen.Context, args json.RawMessage) (microgen.Context, error) {
	cfg := grpcGokitConfig{}
	err := json.Unmarshal(args, &cfg)
	if err != nil {
		return ctx, err
	}
	if cfg.Path == "" {
		cfg.Path = "transport/grpc"
	}

	ctx, err = p.endpoints(ctx, cfg)
	if err != nil {
		return ctx, err
	}
	return ctx, nil
}

func (p *grpcGokitPlugin) endpoints(ctx microgen.Context, cfg grpcGokitConfig) (microgen.Context, error) {
	ImportAliasFromSources = true
	pluginPackagePath, err := gen.GetPkgPath(filepath.Join(cfg.Path, "endpoints.microgen.go"), false)
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

	outfile := microgen.File{}
	outfile.Name = grpcKitPlugin
	outfile.Path = filepath.Join(cfg.Path, "endpoints.microgen.go")
	var b bytes.Buffer
	err = f.Render(&b)
	if err != nil {
		return ctx, err
	}
	outfile.Content = b.Bytes()
	ctx.Files = append(ctx.Files, outfile)
	return ctx, nil
}
