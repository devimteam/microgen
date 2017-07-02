package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"github.com/cv21/microgen/parser"
	astparser "go/parser"
	"go/token"
	"github.com/cv21/microgen/generator"
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

	f, err := astparser.ParseFile(token.NewFileSet(), path, nil, 0)
	if err != nil {
		panic(fmt.Errorf("unable to parse file: %v", err))
	}


	fs, err := parser.GetInterfaceFuncSignatures(f, *flagIfaceName)
	generator.GenerateRequestsFile(fs)
}
