package generator

import (
	"fmt"
	"strings"

	"github.com/devimteam/microgen/generator/template"
	"github.com/vetcher/godecl/types"
)

func Decide(p *types.Interface, force bool, packageName, packagePath string) ([]Template, error) {
	var genTags []string
	for _, comment := range p.Docs {
		if strings.HasPrefix(comment, "//@") {
			genTags = append(genTags, strings.Split(comment[3:], ",")...)
		}
	}

	tmpls := []Template{
		&template.ExchangeTemplate{ServicePackageName: packageName},
		&template.EndpointsTemplate{ServicePackageName: packageName},
	}

	for _, tag := range genTags {
		t := tagToTemplate(tag, p.Methods, packagePath, packageName, force)
		if t == nil {
			return nil, fmt.Errorf("unexpected tag %s", tag)
		}
		tmpls = append(tmpls, t...)
	}
	return tmpls, nil
}

func tagToTemplate(tag string, methods []*types.Function, packagePath, servicePackageName string, force bool) []Template {
	switch tag {
	case "middleware":
		return []Template{&template.MiddlewareTemplate{PackagePath: packagePath}}
	case "logging":
		return []Template{&template.LoggingTemplate{PackagePath: packagePath, IfaceFunctions: methods, Overwrite: force}}
	case "grpc":
		return []Template{
			&template.GRPCClientTemplate{PackagePath: packagePath, ServicePackageName: servicePackageName},
			&template.GRPCServerTemplate{ServicePackageName: servicePackageName, PackagePath: packagePath},
			&template.GRPCEndpointConverterTemplate{PackagePath: packagePath, ServicePackageName: servicePackageName},
			&template.StubGRPCTypeConverterTemplate{PackagePath: packagePath, ServicePackageName: servicePackageName, Methods: methods},
		}
	case "grpc-client":
		return []Template{
			&template.GRPCClientTemplate{PackagePath: packagePath, ServicePackageName: servicePackageName},
			&template.GRPCEndpointConverterTemplate{PackagePath: packagePath, ServicePackageName: servicePackageName},
			&template.StubGRPCTypeConverterTemplate{PackagePath: packagePath, ServicePackageName: servicePackageName, Methods: methods},
		}
	case "grpc-server":
		return []Template{
			&template.GRPCServerTemplate{ServicePackageName: servicePackageName, PackagePath: packagePath},
			&template.GRPCEndpointConverterTemplate{PackagePath: packagePath, ServicePackageName: servicePackageName},
			&template.StubGRPCTypeConverterTemplate{PackagePath: packagePath, ServicePackageName: servicePackageName, Methods: methods},
		}
	case "grpc-conv":
		return []Template{
			&template.GRPCEndpointConverterTemplate{PackagePath: packagePath, ServicePackageName: servicePackageName},
			&template.StubGRPCTypeConverterTemplate{PackagePath: packagePath, ServicePackageName: servicePackageName, Methods: methods},
		}
	}
	return nil
}
