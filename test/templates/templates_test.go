package templates_test

import (
	"bytes"
	"fmt"
	goparser "go/parser"
	"go/token"
	"io/ioutil"
	"testing"

	"github.com/devimteam/microgen/generator"
	"github.com/devimteam/microgen/generator/template"
	parser "github.com/devimteam/microgen/parser"
)

func loadInterface(sourceFile, ifaceName string) (*parser.Interface, error) {
	src, err := ioutil.ReadFile(sourceFile)
	if err != nil {
		return nil, fmt.Errorf("read source file error: %v", err)
	}
	f, err := goparser.ParseFile(token.NewFileSet(), "", src, 0)
	if err != nil {
		return nil, fmt.Errorf("unable to parse file: %v", err)
	}
	fs, err := parser.ParseInterface(f, ifaceName)
	if err != nil {
		return nil, fmt.Errorf("could not get interface func signatures: %v", err)
	}
	return fs, nil
}

func TestTemplates(t *testing.T) {
	allTemplateTests := []struct {
		TestName    string
		Template    generator.Template
		OutFilePath string
	}{
		{
			TestName:    "Endpoints",
			Template:    &template.EndpointsTemplate{},
			OutFilePath: "endpoints.go.txt",
		},
		{
			TestName:    "Exchange",
			Template:    &template.ExchangeTemplate{},
			OutFilePath: "exchange.go.txt",
		},
		{
			TestName:    "Client",
			Template:    &template.ClientTemplate{},
			OutFilePath: "client.go.txt",
		},
		{
			TestName:    "Middleware",
			Template:    &template.MiddlewareTemplate{PackagePath: "github.com/devimteam/microgen/test/svc"},
			OutFilePath: "middleware.go.txt",
		},
		{
			TestName:    "Logging",
			Template:    &template.LoggingTemplate{PackagePath: "github.com/devimteam/microgen/test/svc"},
			OutFilePath: "logging.go.txt",
		},
		{
			TestName:    "GRPC Server",
			Template:    &template.GRPCServerTemplate{},
			OutFilePath: "grpc_server.go.txt",
		},
		{
			TestName:    "GRPC Client",
			Template:    &template.GRPCClientTemplate{PackagePath: "github.com/devimteam/microgen/test/svc"},
			OutFilePath: "grpc_client.go.txt",
		},
		{
			TestName:    "GRPC Converter",
			Template:    &template.GRPCEndpointConverterTemplate{PackagePath: "github.com/devimteam/microgen/test/svc"},
			OutFilePath: "grpc_converters.go.txt",
		},
		{
			TestName:    "GRPC Type Converter",
			Template:    &template.StubGRPCTypeConverterTemplate{PackagePath: "github.com/devimteam/microgen/test/svc"},
			OutFilePath: "grpc_type.go.txt",
		},
	}
	for _, test := range allTemplateTests {
		t.Run(test.TestName, func(t *testing.T) {
			fs, err := loadInterface("service.go.txt", "StringService")
			if err != nil {
				t.Fatal(err)
			}
			out, err := ioutil.ReadFile(test.OutFilePath)
			if err != nil {
				t.Fatalf("read out file error: %v", err)
			}

			buf := bytes.NewBuffer([]byte{})
			gen := generator.NewGenerator([]generator.Template{test.Template}, fs, generator.NewWriterStrategy(buf))
			err = gen.Generate()
			if err != nil {
				t.Errorf("unable to generate: %v", err)
			}
			if buf.String() != string(out[:]) {
				t.Errorf("Got:\n\n%s\n\nExpected:\n\n%s", buf.String(), string(out[:]))
			}
		})
	}
}
