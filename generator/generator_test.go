package generator

import (
	"bytes"
	"fmt"
	goparser "go/parser"
	"go/token"
	"io/ioutil"
	"path/filepath"
	"testing"

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
	// TODO(nicolai): Figure out if packagePath needs to be anything
	packagePath := ""
	info, err := godecl.ParseAstFile(tree, packagePath)
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
	outPath := "./test_out/"
	sourcePath := "./test_assets/service.go.txt"
	absSourcePath, err := filepath.Abs(sourcePath)
	importPackagePath, err := resolvePackagePath(outPath)
	iface, err := loadInterface(sourcePath, "StringService")
	if err != nil {
		t.Fatal(err)
	}

	genInfo := &template.GenerationInfo{
		ServiceImportPackageName: "stringsvc",
		ServiceImportPath:        importPackagePath,
		Force:                    true,
		Iface:                    iface,
		AbsOutPath:               outPath,
		SourceFilePath:           absSourcePath,
		ProtobufPackage:          fetchMetaInfo(TagMark+ProtobufTag, iface.Docs),
		GRPCRegAddr:              fetchMetaInfo(TagMark+GRPCRegAddr, iface.Docs),
	}

	allTemplateTests := []struct {
		TestName    string
		Template    template.Template
		OutFilePath string
	}{
		{
			TestName:    "Endpoints",
			Template:    template.NewEndpointsTemplate(genInfo),
			OutFilePath: "endpoints.go.txt",
		},
		{
			TestName:    "Exchange",
			Template:    template.NewExchangeTemplate(genInfo),
			OutFilePath: "exchange.go.txt",
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
			out, err := ioutil.ReadFile("test_assets/" + test.OutFilePath)
			if err != nil {
				t.Fatalf("read out file error: %v", err)
			}

			buf := bytes.NewBuffer([]byte{})
			// TODO(nicolai): NewGenerator was commented out in [1]
			// Need to use NewGenUnit now.
			// Also consider testify instead of doing diff heavy-lifting
			// [1] https://github.com/devimteam/microgen/commit/0e094fadf97df0da8c14f5b778a3beea4f646545#diff-338bea067d2a01b69c90f1c032c6a24b

			absOutPath := "./test_out/"
			gen, err := NewGenUnit(test.Template, absOutPath)
			if err != nil {
				t.Fatalf("NewGenUnit: %v", err)
			}
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
