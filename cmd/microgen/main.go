package main

import (
	"flag"
	"fmt"
	astparser "go/parser"
	"go/token"
	"os"
	"path/filepath"

	"go/ast"

	"github.com/davecgh/go-spew/spew"
	"github.com/devimteam/microgen/generator"
	"github.com/devimteam/microgen/generator/template"
	"github.com/devimteam/microgen/parser"
)

var (
	flagFileName    = flag.String("file", "", "Name of file where described interface definition")
	flagIfaceName   = flag.String("interface", "", "Interface name")
	flagOutputDir   = flag.String("out", "", "Output directory")
	flagPackagePath = flag.String("package", "", "Service package path")
	flagGRPC        = flag.Bool("grpc", false, "Render gRPC transport")
	flagDebug       = flag.Bool("debug", false, "Debug mode")
	flagHelp        = flag.Bool("help", false, "Show help")
)

func init() {
	flag.Parse()
}

func main() {
	if *flagHelp || *flagFileName == "" || *flagIfaceName == "" || *flagPackagePath == "" {
		flag.Usage()
		os.Exit(0)
	}

	currentDir, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	path := filepath.Join(currentDir, *flagFileName)
	fset := token.NewFileSet()
	f, err := astparser.ParseFile(fset, path, nil, astparser.ParseComments)
	if err != nil {
		panic(fmt.Errorf("error when parse file: %v", err))
	}
	i, err := parser.ParseInterface(f, *flagIfaceName)
	if err != nil {
		panic(fmt.Errorf("error when parse interface from file : %v", err))
	}

	if *flagDebug {
		ast.Print(fset, f)
		spew.Dump(i)
	}

	var strategy generator.Strategy
	if *flagOutputDir == "" {
		strategy = generator.NewWriterStrategy(os.Stdout)
	} else {
		strategy = generator.NewFileStrategy(*flagOutputDir)
	}

	templates := []generator.Template{
		&template.ExchangeTemplate{},
		&template.EndpointsTemplate{},
		&template.ClientTemplate{},
		&template.MiddlewareTemplate{PackagePath: *flagPackagePath},
		&template.LoggingTemplate{PackagePath: *flagPackagePath},
	}
	if *flagGRPC {
		templates = append(templates,
			&template.GRPCServerTemplate{},
			&template.GRPCClientTemplate{PackagePath: *flagPackagePath},
			&template.GRPCConverterTemplate{PackagePath: *flagPackagePath},
		)
	}

	gen := generator.NewGenerator(templates, i, strategy)
	err = gen.Generate()

	if err != nil {
		fmt.Println(err.Error())
	}
}
