package generator

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	mstrings "github.com/devimteam/microgen/generator/strings"
	"github.com/devimteam/microgen/generator/template"
	lg "github.com/devimteam/microgen/logger"
	"github.com/devimteam/microgen/util"
	"github.com/vetcher/godecl/types"
)

const (
	TagMark         = template.TagMark
	MicrogenMainTag = template.MicrogenMainTag
	ProtobufTag     = "protobuf"
	GRPCRegAddr     = "grpc-addr"

	MiddlewareTag             = template.MiddlewareTag
	LoggingMiddlewareTag      = template.LoggingMiddlewareTag
	RecoverMiddlewareTag      = template.RecoverMiddlewareTag
	HttpTag                   = template.HttpTag
	HttpServerTag             = template.HttpServerTag
	HttpClientTag             = template.HttpClientTag
	GrpcTag                   = template.GrpcTag
	GrpcServerTag             = template.GrpcServerTag
	GrpcClientTag             = template.GrpcClientTag
	MainTag                   = template.MainTag
	ErrorLoggingMiddlewareTag = template.ErrorLoggingMiddlewareTag
	TracingTag                = template.TracingTag
	CacheTag                  = template.CacheTag
	JSONRPCTag                = template.JSONRPCTag
	JSONRPCServerTag          = template.JSONRPCServerTag
	JSONRPCClientTag          = template.JSONRPCClientTag

	HttpMethodTag  = template.HttpMethodTag
	HttpMethodPath = template.HttpMethodPath
)

func ListTemplatesForGen(iface *types.Interface, force bool, importPackageName, absOutPath, sourcePath string) (units []*generationUnit, err error) {
	importPackagePath, err := resolvePackagePath(absOutPath)
	if err != nil {
		return nil, err
	}
	absSourcePath, err := filepath.Abs(sourcePath)
	if err != nil {
		return nil, err
	}
	info := &template.GenerationInfo{
		ServiceImportPackageName: importPackageName,
		ServiceImportPath:        importPackagePath,
		Force:                    force,
		Iface:                    iface,
		AbsOutPath:               absOutPath,
		SourceFilePath:           absSourcePath,
		ProtobufPackage:          mstrings.FetchMetaInfo(TagMark+ProtobufTag, iface.Docs),
		GRPCRegAddr:              mstrings.FetchMetaInfo(TagMark+GRPCRegAddr, iface.Docs),
		FileHeader:               defaultFileHeader,
	}
	stubSvc, err := NewGenUnit(template.NewStubInterfaceTemplate(info), absOutPath)
	if err != nil {
		return nil, err
	}
	exch, err := NewGenUnit(template.NewExchangeTemplate(info), absOutPath)
	if err != nil {
		return nil, err
	}
	endp, err := NewGenUnit(template.NewEndpointsTemplate(info), absOutPath)
	if err != nil {
		return nil, err
	}
	units = append(units, stubSvc, exch, endp)

	genTags := util.FetchTags(iface.Docs, TagMark+MicrogenMainTag)
	lg.Logger.Logln(2, "Tags:", strings.Join(genTags, ", "))
	for _, tag := range genTags {
		templates := tagToTemplate(tag, info)
		if templates == nil {
			lg.Logger.Logf(1, "Warning! unexpected tag %s\n", tag)
			continue
		}
		for _, t := range templates {
			unit, err := NewGenUnit(t, absOutPath)
			if err != nil {
				return nil, err
			}
			units = append(units, unit)
		}
	}
	return units, nil
}

func tagToTemplate(tag string, info *template.GenerationInfo) (tmpls []template.Template) {
	switch tag {
	case MiddlewareTag:
		return append(tmpls, template.NewMiddlewareTemplate(info))
	case LoggingMiddlewareTag:
		return append(tmpls, template.NewLoggingTemplate(info))
	case GrpcTag:
		return append(tmpls,
			template.NewGRPCClientTemplate(info),
			template.NewGRPCServerTemplate(info),
			template.NewGRPCEndpointConverterTemplate(info),
			template.NewStubGRPCTypeConverterTemplate(info),
		)
	case GrpcClientTag:
		return append(tmpls,
			template.NewGRPCClientTemplate(info),
			template.NewGRPCEndpointConverterTemplate(info),
			template.NewStubGRPCTypeConverterTemplate(info),
		)
	case GrpcServerTag:
		return append(tmpls,
			template.NewGRPCServerTemplate(info),
			template.NewGRPCEndpointConverterTemplate(info),
			template.NewStubGRPCTypeConverterTemplate(info),
		)
	case HttpTag:
		return append(tmpls,
			template.NewHttpServerTemplate(info),
			template.NewHttpClientTemplate(info),
			template.NewHttpConverterTemplate(info),
		)
	case HttpServerTag:
		return append(tmpls,
			template.NewHttpServerTemplate(info),
			template.NewHttpConverterTemplate(info),
		)
	case HttpClientTag:
		return append(tmpls,
			template.NewHttpClientTemplate(info),
			template.NewHttpConverterTemplate(info),
		)
	case RecoverMiddlewareTag:
		return append(tmpls, template.NewRecoverTemplate(info))
	case MainTag:
		return append(tmpls, template.NewMainTemplate(info))
	case ErrorLoggingMiddlewareTag:
		return append(tmpls, template.NewErrorLoggingTemplate(info))
	case CacheTag:
		return append(tmpls, template.NewCacheMiddlewareTemplate(info))
	case TracingTag:
		return append(tmpls, template.EmptyTemplate{})
	case JSONRPCTag:
		return append(tmpls,
			template.NewJSONRPCEndpointConverterTemplate(info),
			template.NewJSONRPCClientTemplate(info),
			template.NewJSONRPCServerTemplate(info),
		)
	case JSONRPCClientTag:
		return append(tmpls,
			template.NewJSONRPCEndpointConverterTemplate(info),
			template.NewJSONRPCClientTemplate(info),
		)
	case JSONRPCServerTag:
		return append(tmpls,
			template.NewJSONRPCEndpointConverterTemplate(info),
			template.NewJSONRPCServerTemplate(info),
		)
	}
	return nil
}

func resolvePackagePath(outPath string) (string, error) {
	lg.Logger.Logln(3, "try to resolve current package")
	gopath := os.Getenv("GOPATH")
	if gopath == "" {
		return "", fmt.Errorf("GOPATH is empty")
	}

	absOutPath, err := filepath.Abs(outPath)
	if err != nil {
		return "", err
	}

	gopathSrc := filepath.Join(gopath, "src")
	if !strings.HasPrefix(absOutPath, gopathSrc) {
		return "", fmt.Errorf("path not in GOPATH")
	}

	return absOutPath[len(gopathSrc)+1:], nil
}
