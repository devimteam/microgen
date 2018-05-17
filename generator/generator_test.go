package generator

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"
	"testing"

	"context"

	strings2 "github.com/devimteam/microgen/generator/strings"
	"github.com/devimteam/microgen/generator/template"
	"github.com/stretchr/testify/assert"
	"github.com/vetcher/go-astra"
	"github.com/vetcher/go-astra/types"
)

func findInterface(file *types.File, ifaceName string) *types.Interface {
	for i := range file.Interfaces {
		if file.Interfaces[i].Name == ifaceName {
			return &file.Interfaces[i]
		}
	}
	return nil
}

func loadInterface(sourceFile, ifaceName string) (*types.Interface, error) {
	info, err := astra.ParseFile(sourceFile)
	if err != nil {
		return nil, err
	}
	i := findInterface(info, ifaceName)
	if i == nil {
		return nil, fmt.Errorf("could not find %s interface", ifaceName)
	}
	return i, nil
}

func TestTemplates(t *testing.T) {
	outPath := "./test_out/"
	sourcePath := "./test_assets/service.go.txt"
	absSourcePath, err := filepath.Abs(sourcePath)
	importPackagePath, err := resolvePackagePath(outPath)
	iface, err := loadInterface(sourcePath, "StringService")
	if err != nil {
		t.Fatal(err)
	}

	genInfo := &template.GenerationInfo{
		SourcePackageImport:   importPackagePath,
		Iface:                 iface,
		OutputFilePath:        outPath,
		SourceFilePath:        absSourcePath,
		ProtobufPackageImport: strings2.FetchMetaInfo(TagMark+ProtobufTag, iface.Docs),
	}
	t.Log("protobuf pkg", genInfo.ProtobufPackageImport)

	allTemplateTests := []struct {
		TestName    string
		Template    template.Template
		OutFilePath string
	}{
		{
			TestName:    "Endpoints",
			Template:    template.NewEndpointsTemplate(genInfo),
			OutFilePath: "transport_endpoints.go.txt",
		},
		{
			TestName:    "Exchange",
			Template:    template.NewExchangeTemplate(genInfo),
			OutFilePath: "transport_exchanges.go.txt",
		},
		{
			TestName:    "Middleware",
			Template:    template.NewMiddlewareTemplate(genInfo),
			OutFilePath: "middleware.go.txt",
		},
		{
			TestName:    "Logging",
			Template:    template.NewLoggingTemplate(genInfo),
			OutFilePath: "logging.go.txt",
		},
		{
			TestName:    "GRPC Server",
			Template:    template.NewGRPCServerTemplate(genInfo),
			OutFilePath: "grpc_server.go.txt",
		},
		{
			TestName:    "GRPC Client",
			Template:    template.NewGRPCClientTemplate(genInfo),
			OutFilePath: "grpc_client.go.txt",
		},
		{
			TestName:    "GRPC Converter",
			Template:    template.NewGRPCEndpointConverterTemplate(genInfo),
			OutFilePath: "grpc_converters.go.txt",
		},
		{
			TestName:    "GRPC Type Converter",
			Template:    template.NewStubGRPCTypeConverterTemplate(genInfo),
			OutFilePath: "grpc_type.go.txt",
		},
	}
	for _, test := range allTemplateTests {
		t.Run(test.TestName, func(t *testing.T) {
			expected, err := ioutil.ReadFile("test_assets/" + test.OutFilePath)
			if err != nil {
				t.Fatalf("read expected file error: %v", err)
			}

			absOutPath := "./test_out/"
			gen, err := NewGenUnit(context.Background(), test.Template, absOutPath)
			if err != nil {
				t.Fatalf("NewGenUnit: %v", err)
			}
			err = gen.Generate(context.Background())
			if err != nil {
				t.Fatalf("unable to generate: %v", err)
			}
			actual, err := ioutil.ReadFile("./test_out/" + test.Template.DefaultPath())
			if err != nil {
				t.Fatalf("read actual file error: %v", err)
			}
			assert.Equal(t,
				strings.Split(string(expected[:]), "\n"),
				strings.Split(string(actual[:]), "\n"),
			)
		})
	}
}
