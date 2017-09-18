package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/devimteam/microgen/generator"
	"github.com/devimteam/microgen/util"
	"github.com/vetcher/godecl/types"
)

var (
	flagFileName  = flag.String("file", "service.go", "Name of file where described interface definition")
	flagOutputDir = flag.String("out", ".", "Output directory")
	flagHelp      = flag.Bool("help", false, "Show help")
	flagInit      = flag.Bool("init", false, "Generate stub methods for converters")
)

func init() {
	flag.Parse()
}

func main() {
	if *flagHelp || *flagFileName == "" {
		flag.Usage()
		os.Exit(0)
	}

	info, err := util.ParseFile(*flagFileName)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	i := findInterface(info)
	if i == nil {
		fmt.Println("could not find interface with @microgen tag")
		os.Exit(1)
	}

	if err := generator.ValidateInterface(i); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	var strategy generator.Strategy
	if *flagOutputDir == "" {
		strategy = generator.NewWriterStrategy(os.Stdout)
	} else {
		strategy = generator.NewFileStrategy(*flagOutputDir)
	}

	packagePath, err := resolvePackagePath(*flagOutputDir)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	templates, err := generator.ListTemplatesForGen(i, *flagInit, info.Name, packagePath)

	gen := generator.NewGenerator(templates, i, strategy)
	err = gen.Generate()
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Printf("All files for %s successfully generated\n", i.Name)
	}
}

func resolvePackagePath(outPath string) (string, error) {
	gopath := os.Getenv("GOPATH")
	if gopath == "" {
		return "", fmt.Errorf("GOPATH is empty")
	}

	absOutPath, err := filepath.Abs(outPath)
	if err != nil {
		return "", err
	}

	gopathSrc := filepath.Join(gopath, "src")
	if !strings.HasPrefix(absOutPath, gopathSrc) {
		return "", fmt.Errorf("-out not in GOPATH")
	}

	return absOutPath[len(gopathSrc)+1:], nil
}

func findInterface(file *types.File) *types.Interface {
	for i := range file.Interfaces {
		if docsContainMicrogenTag(file.Interfaces[i].Docs) {
			return &file.Interfaces[i]
		}
	}
	return nil
}

func docsContainMicrogenTag(strs []string) bool {
	for _, str := range strs {
		if strings.HasPrefix(str, generator.MicrogenTag) {
			return true
		}
	}
	return false
}
