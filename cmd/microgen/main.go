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
	flagFileName  = flag.String("file", "", "File name")
	flagIfaceName = flag.String("interface", "", "Interface name")
	flagOutputDir = flag.String("out", "", "Output directory")
	debug         = flag.Bool("debug", false, "Debug mode")
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
		panic(fmt.Errorf("error when parse file: %v", err))
	}
	i, err := parser.ParseInterface(f, *flagIfaceName)
	if err != nil {
		panic(fmt.Errorf("error when parse interface from file : %v", err))
	}

	if *debug {
		ast.Print(fset, f)
		spew.Dump(i)
	}

	gen := generator.NewGenerator([]generator.Template{
		&template.ExchangeTemplate{},
		&template.EndpointsTemplate{},
	}, i, *flagOutputDir)

	err = gen.Generate()

	if err != nil {
		fmt.Println(err.Error())
	}
}
