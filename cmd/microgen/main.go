package main

import (
	"flag"
	"fmt"
	"go/ast"
	astparser "go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"

	"github.com/davecgh/go-spew/spew"
	"github.com/devimteam/microgen/generator"
	"github.com/devimteam/microgen/generator/template"
	"github.com/devimteam/microgen/parser"
)

var (
	flagFileName  = flag.String("file", "service.go", "Name of file where described interface definition")
	flagIfaceName = flag.String("interface", "", "Interface name")
	flagOutputDir = flag.String("out", "", "Output directory")
	flagGRPC      = flag.Bool("grpc", false, "Render gRPC transport")
	flagDebug     = flag.Bool("debug", false, "Debug mode")
	flagHelp      = flag.Bool("help", false, "Show help")
	flagInit      = flag.Bool("init", false, "With flag `-grpc` generate stub methods for converters")
)

func init() {
	flag.Parse()
}

func main() {
	if *flagHelp || *flagFileName == "" || *flagIfaceName == "" {
		flag.Usage()
		os.Exit(0)
	}

	currentDir, err := os.Getwd()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	path := filepath.Join(currentDir, *flagFileName)
	fset := token.NewFileSet()
	f, err := astparser.ParseFile(fset, path, nil, astparser.ParseComments)
	if err != nil {
		fmt.Printf("error when parse file: %v\n", err)
		os.Exit(1)
	}
	i, err := parser.ParseInterface(f, *flagIfaceName)
	if err != nil {
		fmt.Printf("error when parse interface from file : %v\n", err)
		os.Exit(1)
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

	packagePath := resolvePackagePath(*flagOutputDir)
	templates := []generator.Template{
		&template.ExchangeTemplate{},
		&template.EndpointsTemplate{},
		&template.ClientTemplate{},
		&template.MiddlewareTemplate{PackagePath: packagePath},
		&template.LoggingTemplate{PackagePath: packagePath},
	}
	if *flagGRPC {
		templates = append(templates,
			&template.GRPCServerTemplate{},
			&template.GRPCClientTemplate{PackagePath: packagePath},
			&template.GRPCEndpointConverterTemplate{PackagePath: packagePath},
		)
		if *flagInit {
			templates = append(templates,
				&template.StubGRPCTypeConverterTemplate{PackagePath: packagePath},
			)
		}
	}

	gen := generator.NewGenerator(templates, i, strategy)
	err = gen.Generate()
	if err != nil {
		fmt.Println(err.Error())
	}
}

func resolvePackagePath(outPath string) string {
	gopath := os.Getenv("GOPATH")
	if gopath == "" {
		fmt.Println("GOPATH is empty")
		os.Exit(1)
	}

	absOutPath, err := filepath.Abs(outPath)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	gopathSrc := filepath.Join(gopath, "src")
	if !strings.HasPrefix(absOutPath, gopathSrc) {
		fmt.Println("-out not in GOPATH")
		os.Exit(1)
	}

	return absOutPath[len(gopathSrc)+1:]
}
