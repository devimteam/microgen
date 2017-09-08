package generator

import (
	"fmt"
	"strings"

	"github.com/devimteam/microgen/generator/template"
	"github.com/vetcher/godecl/types"
)

func Decide(p *types.Interface, init bool, packageName, packagePath string) ([]Template, error) {
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
		t := tagToTemplate(tag, packagePath, packageName, init)
		if t == nil {
			return nil, fmt.Errorf("unexpected tag %s", tag)
		}
		tmpls = append(tmpls, t...)
	}
	return tmpls, nil
}

func tagToTemplate(tag string, packagePath, servicePackageName string, init bool) (tmpls []Template) {
	switch tag {
	case "middleware":
		return append(tmpls, &template.MiddlewareTemplate{PackagePath: packagePath})
	case "logging":
		return []Template{&template.LoggingTemplate{PackagePath: packagePath}}
	case "grpc":
		if init {
			tmpls = append(tmpls, &template.StubGRPCTypeConverterTemplate{PackagePath: packagePath, ServicePackageName: servicePackageName})
		}
		return append(tmpls,
			&template.GRPCClientTemplate{PackagePath: packagePath, ServicePackageName: servicePackageName},
			&template.GRPCServerTemplate{ServicePackageName: servicePackageName, PackagePath: packagePath},
			&template.GRPCEndpointConverterTemplate{PackagePath: packagePath, ServicePackageName: servicePackageName},
		)
	case "grpc-client":
		if init {
			tmpls = append(tmpls, &template.StubGRPCTypeConverterTemplate{PackagePath: packagePath, ServicePackageName: servicePackageName})
		}
		return append(tmpls,
			&template.GRPCClientTemplate{PackagePath: packagePath, ServicePackageName: servicePackageName},
			&template.GRPCEndpointConverterTemplate{PackagePath: packagePath, ServicePackageName: servicePackageName},
		)
	case "grpc-server":
		if init {
			tmpls = append(tmpls, &template.StubGRPCTypeConverterTemplate{PackagePath: packagePath, ServicePackageName: servicePackageName})
		}
		return append(tmpls,
			&template.GRPCServerTemplate{ServicePackageName: servicePackageName, PackagePath: packagePath},
			&template.GRPCEndpointConverterTemplate{PackagePath: packagePath, ServicePackageName: servicePackageName},
		)
	case "grpc-conv":
		if init {
			tmpls = append(tmpls, &template.StubGRPCTypeConverterTemplate{PackagePath: packagePath, ServicePackageName: servicePackageName})
		}
		return append(tmpls,
			&template.GRPCEndpointConverterTemplate{PackagePath: packagePath, ServicePackageName: servicePackageName},
		)
	}
	return nil
}
