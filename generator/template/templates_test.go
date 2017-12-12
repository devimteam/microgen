package template

import (
	"bytes"
	"fmt"
	goparser "go/parser"
	"go/token"
	"io/ioutil"
	"testing"

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
		Template    Template
		OutFilePath string
	}{
		{
			TestName:    "Endpoints",
			Template:    &endpointsTemplate{ServicePackageName: "stringsvc"},
			OutFilePath: "endpoints.go.txt",
		},
		{
			TestName:    "Exchange",
			Template:    &exchangeTemplate{ServicePackageName: "stringsvc"},
			OutFilePath: "exchange.go.txt",
		},
		{
			TestName:    "Middleware",
			Template:    &middlewareTemplate{PackagePath: "github.com/devimteam/microgen/example/svc"},
			OutFilePath: "middleware.go.txt",
		},
		{
			TestName:    "Logging",
			Template:    &loggingTemplate{PackagePath: "github.com/devimteam/microgen/example/svc"},
			OutFilePath: "logging.go.txt",
		},
		{
			TestName:    "GRPC Server",
			Template:    &gRPCServerTemplate{ServicePackageName: "stringsvc"},
			OutFilePath: "grpc_server.go.txt",
		},
		{
			TestName:    "GRPC Client",
			Template:    &gRPCClientTemplate{PackagePath: "github.com/devimteam/microgen/example/svc"},
			OutFilePath: "grpc_client.go.txt",
		},
		{
			TestName:    "GRPC Converter",
			Template:    &gRPCEndpointConverterTemplate{PackagePath: "github.com/devimteam/microgen/example/svc", ServicePackageName: "stringsvc"},
			OutFilePath: "grpc_converters.go.txt",
		},
		{
			TestName:    "GRPC Type Converter",
			Template:    &stubGRPCTypeConverterTemplate{PackagePath: "github.com/devimteam/microgen/example/svc"},
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
			gen := NewGenerator([]Template{test.Template}, fs, WriterStrategy(buf))
			err = gen.Generate()
			if err != nil {
				t.Fatalf("unable to generate: %v", err)
			}
			if buf.String() != string(out[:]) {
				t.Errorf("Got:\n/////////\n%s\n/////////\nExpected:\n/////////\n%s\n/////////", buf.String(), string(out[:]))
				t.Errorf("1: Got(bytes), 2: Expected(bytes):\n/////////\n1: %v\n2: %v\n/////////", buf.Bytes(), out[:])
				x, y, z, line := findDifference(buf.String(), string(out[:]))
				t.Errorf("line:pos:raw %d:%d:%d %d!=%d %v\n`%s`", x+1, y+1, z, buf.Bytes()[z], out[z], len(buf.String()) == len(string(out[:])), line)
			}
		})
	}
}

func findDifference(first, second string) (line int, pos int, raw int, strLine string) {
	for i, sym := range first {
		if first[i] != second[i] {
			return
		}
		if sym == '\n' {
			line += 1
			pos = 0
			strLine = ""
		}
		pos += 1
		raw += 1
		strLine += string(sym)
	}
	return 0, 0, 0, ""
}
