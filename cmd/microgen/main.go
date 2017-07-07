package main

import (
	"flag"
	"fmt"
	astparser "go/parser"
	"go/token"
	"os"
	"path/filepath"

	"go/ast"

	"github.com/cv21/microgen/generator"
	"github.com/cv21/microgen/parser"
)

var (
	flagFileName  = flag.String("f", "", "File name")
	flagIfaceName = flag.String("i", "", "Interface name")
	debug         = flag.Bool("d", false, "Debug mode")
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

	if *debug {
		ast.Print(fset, f)
	}

	i, err := parser.ParseInterface(f, *flagIfaceName)
	gen := generator.NewGenerator([]*generator.Template{
		&generator.RequestsTemplate,
		&generator.ResponsesTemplate,
		&generator.EndpointTemplate,
	}, i)

	err = gen.Generate()

	fmt.Println(err)
}
