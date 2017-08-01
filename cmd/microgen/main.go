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
	"github.com/devimteam/microgen/parser"
)

var (
	flagFileName    = flag.String("f", "", "File name")
	flagIfaceName   = flag.String("i", "", "Interface name")
	flagOutputDir   = flag.String("o", "", "Output directory")
	flagPkgFullName = flag.String("p", "", "Package full name")
	debug           = flag.Bool("d", false, "Debug mode")
)

func init() {
	flag.Parse()
}

func main() {
	currentDir, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	path := filepath.Join(currentDir, *flagFileName)

	fset := token.NewFileSet()
	f, err := astparser.ParseFile(fset, path, nil, astparser.ParseComments)
	if err != nil {
		panic(fmt.Errorf("unable to parse file: %v", err))
	}
	i, err := parser.ParseInterface(f, *flagIfaceName)

	if *debug {
		ast.Print(fset, f)
		spew.Dump(i)
	}

	gen := generator.NewGenerator([]*generator.Template{
		&generator.ExchangeTemplate,
		&generator.EndpointTemplate,
		&generator.ClientTemplate,
		&generator.LoggingMiddlewareTemplate,
	}, generator.RenderData{
		Interface:       i,
		PackageFullName: *flagPkgFullName,
	}, *flagOutputDir)

	err = gen.Generate()

	if err != nil {
		fmt.Println(err.Error())
	}
}
