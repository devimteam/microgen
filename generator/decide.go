package generator

import (
	"fmt"
	"strings"

	"os"
	"path/filepath"

	"github.com/devimteam/microgen/generator/template"
	"github.com/devimteam/microgen/util"
	"github.com/vetcher/godecl/types"
)

const (
	MicrogenGeneralTag = "// @microgen"
	ProtobufTag        = "// @protobuf"
	GRPCRegAddr        = "// @grpc-addr"
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
		ProtobufPackage:          fetchMetaInfo(ProtobufTag, iface.Docs),
		GRPCRegAddr:              fetchMetaInfo(GRPCRegAddr, iface.Docs),
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

	genTags := util.FetchTags(iface.Docs, MicrogenGeneralTag)
	for _, tag := range genTags {
		templates := tagToTemplate(tag, info)
		if templates == nil {
			return nil, fmt.Errorf("unexpected tag %s", tag)
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

// Fetch information from slice of comments (docs).
// Returns appendix of first comment which has tag as prefix.
func fetchMetaInfo(tag string, comments []string) string {
	for _, comment := range comments {
		if len(comment) > len(tag) && strings.HasPrefix(comment, tag) {
			return comment[len(tag)+1:]
		}
	}
	return ""
}

func tagToTemplate(tag string, info *template.GenerationInfo) (tmpls []template.Template) {
	switch tag {
	case "middleware":
		return []template.Template{template.NewMiddlewareTemplate(info)}
	case "logging":
		return []template.Template{template.NewLoggingTemplate(info)}
	case "grpc":
		return append(tmpls,
			template.NewGRPCClientTemplate(info),
			template.NewGRPCServerTemplate(info),
			template.NewGRPCEndpointConverterTemplate(info),
			template.NewStubGRPCTypeConverterTemplate(info),
		)
	case "grpc-client":
		return append(tmpls,
			template.NewGRPCClientTemplate(info),
			template.NewGRPCEndpointConverterTemplate(info),
			template.NewStubGRPCTypeConverterTemplate(info),
		)
	case "grpc-server":
		return append(tmpls,
			template.NewGRPCServerTemplate(info),
			template.NewGRPCEndpointConverterTemplate(info),
			template.NewStubGRPCTypeConverterTemplate(info),
		)
	}
	return nil
}

func resolvePackagePath(outPath string) (string, error) {
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
