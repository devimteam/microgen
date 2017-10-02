package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/devimteam/microgen/generator"
	"github.com/devimteam/microgen/util"
	"github.com/vetcher/godecl/types"
)

var (
	flagFileName  = flag.String("file", "service.go", "Name of file where described interface definition")
	flagOutputDir = flag.String("out", ".", "Output directory")
	flagHelp      = flag.Bool("help", false, "Show help")
	flagForce     = flag.Bool("force", false, "Overwrite all files, as it generates for the first time")
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
		fmt.Println("fatal:", err)
		os.Exit(1)
	}

	i := findInterface(info)
	if i == nil {
		fmt.Println("fatal: could not find interface with @microgen tag")
		os.Exit(1)
	}

	if err := generator.ValidateInterface(i); err != nil {
		fmt.Println("validation:", err)
		os.Exit(1)
	}

	units, err := generator.ListTemplatesForGen(i, *flagForce, info.Name, *flagOutputDir, *flagFileName)
	if err != nil {
		fmt.Println("fatal:", err)
		os.Exit(1)
	}
	for _, unit := range units {
		err := unit.Generate()
		if err != nil {
			fmt.Println("fatal:", err)
			os.Exit(1)
		}
	}
	fmt.Println("All files successfully generated")
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
		if strings.HasPrefix(str, generator.MicrogenGeneralTag) {
			return true
		}
	}
	return false
}
