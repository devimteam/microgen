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

	info, err := util.ParseFile(*flagFileName)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	i := findInterface(info, *flagIfaceName)
	if i == nil {
		fmt.Printf("could not find %s interface", *flagIfaceName)
		os.Exit(1)
	}

	if err := generator.ValidateInterface(i); err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	var strategy generator.Strategy
	if *flagOutputDir == "" {
		strategy = generator.WriterStrategy(os.Stdout)
	} else {
		strategy = generator.NewFileStrategy(*flagOutputDir)
	}

	packagePath := resolvePackagePath(*flagOutputDir)
	templates, err := generator.Decide(i, true, info.Name, packagePath)

	gen := generator.NewForceGenerator(templates, i, strategy)
	err = gen.Generate()
	if err != nil {
		fmt.Println(err.Error())
	} else {
		fmt.Println("All files successfully generated")
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

func findInterface(file *types.File, ifaceName string) *types.Interface {
	for i := range file.Interfaces {
		if file.Interfaces[i].Name == ifaceName {
			return &file.Interfaces[i]
		}
	}
	return nil
}
