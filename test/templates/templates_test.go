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
	"github.com/vetcher/godecl"
	"github.com/vetcher/godecl/types"
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
	src, err := ioutil.ReadFile(sourceFile)
	if err != nil {
		return nil, fmt.Errorf("read source file error: %v", err)
	}
	tree, err := goparser.ParseFile(token.NewFileSet(), "", src, 0)
	if err != nil {
		return nil, fmt.Errorf("unable to parse file: %v", err)
	}
	info, err := godecl.ParseFile(tree)
	if err != nil {
		fmt.Printf("error when parsing info from file: %v\n", err)
	}
	i := findInterface(info, ifaceName)
	if i == nil {
		return nil, fmt.Errorf("could not find %s interface", ifaceName)
	}
	return i, nil
}

func TestTemplates(t *testing.T) {
	fs, err := loadInterface("service.go.txt", "StringService")
	if err != nil {
		t.Fatal(err)
	}
	allTemplateTests := []struct {
		TestName    string
		Template    generator.Template
		OutFilePath string
	}{
		{
			TestName:    "Endpoints",
			Template:    &template.EndpointsTemplate{ServicePackageName: "stringsvc"},
			OutFilePath: "endpoints.go.txt",
		},
		{
			TestName:    "Exchange",
			Template:    &template.ExchangeTemplate{ServicePackageName: "stringsvc"},
			OutFilePath: "exchange.go.txt",
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
			Template:    &template.GRPCServerTemplate{ServicePackageName: "stringsvc"},
			OutFilePath: "grpc_server.go.txt",
		},
		{
			TestName:    "GRPC Client",
			Template:    &template.GRPCClientTemplate{PackagePath: "github.com/devimteam/microgen/test/svc"},
			OutFilePath: "grpc_client.go.txt",
		},
		{
			TestName:    "GRPC Converter",
			Template:    &template.GRPCEndpointConverterTemplate{PackagePath: "github.com/devimteam/microgen/test/svc", ServicePackageName: "stringsvc"},
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
			out, err := ioutil.ReadFile(test.OutFilePath)
			if err != nil {
				t.Fatalf("read out file error: %v", err)
			}

			buf := bytes.NewBuffer([]byte{})
			gen := generator.NewGenerator([]generator.Template{test.Template}, fs, generator.WriterStrategy(buf))
			err = gen.Generate()
			if err != nil {
				t.Fatalf("unable to generate: %v", err)
			}
			if buf.String() != string(out[:]) {
				t.Errorf("Got:\n/////////\n%s\n/////////\nExpected:\n/////////\n%s\n/////////", buf.String(), string(out[:]))
				t.Errorf("1: Got(bytes), 2: Expected(bytes):\n/////////\n1: %v\n2: %v\n/////////", buf.Bytes(), out[:])
				x, y, _ := findDifference(buf.String(), string(out[:]))
				t.Errorf("%d:%d", x, y)
			}
		})
	}
}

func findDifference(first, second string) (line int, pos int, raw int) {
	for i, sym := range first {
		if first[i] != second[i] {
			return
		}
		if sym == '\n' {
			line += 1
			pos = 0
		}
		pos += 1
		raw += 1
	}
	return 0, 0, 0
}
