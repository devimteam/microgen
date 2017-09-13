package generator

import (
	"fmt"
	"strings"

	"os"
	"path/filepath"

	"github.com/devimteam/microgen/generator/template"
	"github.com/vetcher/godecl/types"
)

func Decide(iface *types.Interface, force bool, importPackageName, absOutPath string) (units []*generationUnit, err error) {
	importPackagePath, err := resolvePackagePath(absOutPath)
	if err != nil {
		return nil, err
	}
	info := &template.GenerationInfo{
		ServiceImportPackageName: importPackageName,
		ServiceImportPath:        importPackagePath,
		Force:                    force,
		Iface:                    iface,
		AbsOutPath:               absOutPath,
	}

	exch, err := NewGenUnit(template.NewExchangeTemplate(info), info, absOutPath)
	endp, err := NewGenUnit(template.NewEndpointsTemplate(info), info, absOutPath)
	units = append(units, exch, endp)

	genTags := fetchTags(iface.Docs)
	for _, tag := range genTags {
		templates := tagToTemplate(tag, info)
		if templates == nil {
			return nil, fmt.Errorf("unexpected tag %s", tag)
		}
		for _, t := range templates {
			unit, err := NewGenUnit(t, info, absOutPath)
			if err != nil {
				return nil, err
			}
			units = append(units, unit)
		}
	}
	return units, nil
}

func tagToTemplate(tag string, info *template.GenerationInfo) (tmpls []Template) {
	switch tag {
	case "middleware":
		return []Template{template.NewMiddlewareTemplate(info)}
	case "logging":
		return []Template{template.NewLoggingTemplate(info)}
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

func fetchTags(strs []string) (tags []string) {
	for _, comment := range strs {
		if strings.HasPrefix(comment, "//@") {
			tags = append(tags, strings.Split(strings.Replace(comment[3:], " ", "", -1), ",")...)
		}
	}
	return
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
