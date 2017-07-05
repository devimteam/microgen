package main

import (
	"flag"
	"fmt"
	astparser "go/parser"
	"go/token"
	"os"
	"path/filepath"

	"github.com/cv21/microgen/generator"
	"github.com/cv21/microgen/parser"
)

var (
	flagFileName  = flag.String("f", "", "File name")
	flagIfaceName = flag.String("i", "", "Interface name")
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

	f, err := astparser.ParseFile(token.NewFileSet(), path, nil, astparser.ParseComments)
	if err != nil {
		panic(fmt.Errorf("unable to parse file: %v", err))
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
