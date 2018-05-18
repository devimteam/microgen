package generator

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	mstrings "github.com/devimteam/microgen/generator/strings"
	"github.com/devimteam/microgen/generator/template"
	lg "github.com/devimteam/microgen/logger"
	"github.com/vetcher/go-astra/types"
)

const (
	TagMark         = template.TagMark
	MicrogenMainTag = template.MicrogenMainTag
	ProtobufTag     = "protobuf"
	GRPCRegAddr     = "grpc-addr"

	MiddlewareTag             = template.MiddlewareTag
	LoggingMiddlewareTag      = template.LoggingMiddlewareTag
	RecoveringMiddlewareTag   = template.RecoveringMiddlewareTag
	HttpTag                   = template.HttpTag
	HttpServerTag             = template.HttpServerTag
	HttpClientTag             = template.HttpClientTag
	GrpcTag                   = template.GrpcTag
	GrpcServerTag             = template.GrpcServerTag
	GrpcClientTag             = template.GrpcClientTag
	MainTag                   = template.MainTag
	ErrorLoggingMiddlewareTag = template.ErrorLoggingMiddlewareTag
	TracingMiddlewareTag      = template.TracingMiddlewareTag
	CachingMiddlewareTag      = template.CachingMiddlewareTag
	JSONRPCTag                = template.JSONRPCTag
	JSONRPCServerTag          = template.JSONRPCServerTag
	JSONRPCClientTag          = template.JSONRPCClientTag
	Transport                 = template.Transport
	TransportClient           = template.TransportClient
	TransportServer           = template.TransportServer
	MetricsMiddlewareTag      = template.MetricsMiddlewareTag
	ServiceDiscoveryTag       = template.ServiceDiscoveryTag

	HttpMethodTag  = template.HttpMethodTag
	HttpMethodPath = template.HttpMethodPath
)

func ListTemplatesForGen(ctx context.Context, iface *types.Interface, absOutPath, sourcePath string) (units []*GenerationUnit, err error) {
	importPackagePath, err := resolvePackagePath(sourcePath)
	if err != nil {
		return nil, err
	}
	absSourcePath, err := filepath.Abs(sourcePath)
	if err != nil {
		return nil, err
	}
	outImportPath, err := resolvePackagePath(absOutPath)
	if err != nil {
		return nil, err
	}
	m := make(map[string]bool, len(iface.Methods))
	for _, fn := range iface.Methods {
		m[fn.Name] = !mstrings.ContainTag(mstrings.FetchTags(fn.Docs, TagMark+MicrogenMainTag), "-")
	}
	info := &template.GenerationInfo{
		SourcePackageImport:   importPackagePath,
		SourceFilePath:        absSourcePath,
		Iface:                 iface,
		OutputPackageImport:   outImportPath,
		OutputFilePath:        absOutPath,
		ProtobufPackageImport: mstrings.FetchMetaInfo(TagMark+ProtobufTag, iface.Docs),
		FileHeader:            defaultFileHeader,
		AllowedMethods:        m,
	}
	/*stubSvc, err := NewGenUnit(ctx, template.NewStubInterfaceTemplate(info), absOutPath)
	if err != nil {
		return nil, err
	}
	units = append(units, stubSvc)*/

	genTags := mstrings.FetchTags(iface.Docs, TagMark+MicrogenMainTag)
	lg.Logger.Logln(2, "Tags:", strings.Join(genTags, ", "))
	uniqueTemplate := make(map[string]template.Template)
	for _, tag := range genTags {
		templates := tagToTemplate(tag, info)
		if templates == nil {
			lg.Logger.Logf(1, "Warning! unexpected tag %s\n", tag)
			continue
		}
		for _, t := range templates {
			uniqueTemplate[t.DefaultPath()] = t
		}
	}
	for _, t := range uniqueTemplate {
		unit, err := NewGenUnit(ctx, t, absOutPath)
		if err != nil {
			return nil, fmt.Errorf("%s: %v", absOutPath, err)
		}
		units = append(units, unit)
	}
	return units, nil
}

func tagToTemplate(tag string, info *template.GenerationInfo) (tmpls []template.Template) {
	switch tag {
	case MiddlewareTag:
		return append(tmpls, template.NewMiddlewareTemplate(info))
	case LoggingMiddlewareTag:
		return append(
			append(tmpls, tagToTemplate(MiddlewareTag, info)...),
			template.NewLoggingTemplate(info),
		)
	case GrpcTag:
		return append(
			append(tmpls, tagToTemplate(Transport, info)...),
			template.NewGRPCClientTemplate(info),
			template.NewGRPCServerTemplate(info),
			template.NewGRPCEndpointConverterTemplate(info),
			template.NewStubGRPCTypeConverterTemplate(info),
		)
	case GrpcClientTag:
		return append(
			append(tmpls, tagToTemplate(TransportClient, info)...),
			template.NewGRPCClientTemplate(info),
			template.NewGRPCEndpointConverterTemplate(info),
			template.NewStubGRPCTypeConverterTemplate(info),
		)
	case GrpcServerTag:
		return append(
			append(tmpls, tagToTemplate(TransportServer, info)...),
			template.NewGRPCServerTemplate(info),
			template.NewGRPCEndpointConverterTemplate(info),
			template.NewStubGRPCTypeConverterTemplate(info),
		)
	case HttpTag:
		return append(
			append(tmpls, tagToTemplate(Transport, info)...),
			template.NewHttpServerTemplate(info),
			template.NewHttpClientTemplate(info),
			template.NewHttpConverterTemplate(info),
		)
	case HttpServerTag:
		return append(
			append(tmpls, tagToTemplate(TransportServer, info)...),
			template.NewHttpServerTemplate(info),
			template.NewHttpConverterTemplate(info),
		)
	case HttpClientTag:
		return append(
			append(tmpls, tagToTemplate(TransportClient, info)...),
			template.NewHttpClientTemplate(info),
			template.NewHttpConverterTemplate(info),
		)
	case RecoveringMiddlewareTag:
		return append(
			append(tmpls, tagToTemplate(MiddlewareTag, info)...),
			template.NewRecoverTemplate(info),
		)
	case MainTag:
		return append(tmpls, template.NewMainTemplate(info))
	case ErrorLoggingMiddlewareTag:
		return append(
			append(tmpls, tagToTemplate(MiddlewareTag, info)...),
			template.NewErrorLoggingTemplate(info),
		)
	case CachingMiddlewareTag:
		return append(
			append(tmpls, tagToTemplate(MiddlewareTag, info)...),
			template.NewCacheMiddlewareTemplate(info),
		)
	case TracingMiddlewareTag:
		return append(tmpls, template.EmptyTemplate{})
	case MetricsMiddlewareTag:
		return append(tmpls, template.EmptyTemplate{})
	case ServiceDiscoveryTag:
		return append(tmpls, template.EmptyTemplate{})
	case Transport:
		return append(tmpls,
			template.NewExchangeTemplate(info),
			template.NewEndpointsTemplate(info),
			template.NewEndpointsClientTemplate(info),
			template.NewEndpointsServerTemplate(info),
		)
	case TransportClient:
		return append(tmpls,
			template.NewExchangeTemplate(info),
			template.NewEndpointsTemplate(info),
			template.NewEndpointsClientTemplate(info),
		)
	case TransportServer:
		return append(tmpls,
			template.NewExchangeTemplate(info),
			template.NewEndpointsTemplate(info),
			template.NewEndpointsServerTemplate(info),
		)
		// JSON-RPC commented for now, and, I think, will be deleted in future.
		/*case JSONRPCTag:
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
			)*/
	}
	return nil
}

func resolvePackagePath(outPath string) (string, error) {
	lg.Logger.Logln(3, "try to resolve current package")
	gopath := os.Getenv("GOPATH")
	if gopath == "" {
		return "", fmt.Errorf("GOPATH is empty")
	}

	outPath = filepath.Dir(outPath)
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
