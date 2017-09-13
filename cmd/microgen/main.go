package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/devimteam/microgen/generator"
	"github.com/devimteam/microgen/util"
	"github.com/vetcher/godecl/types"
)

var (
	flagFileName  = flag.String("file", "service.go", "Name of file where described interface definition")
	flagIfaceName = flag.String("interface", "", "Interface name")
	flagOutputDir = flag.String("out", "", "Output directory")
	flagHelp      = flag.Bool("help", false, "Show help")
	flagInit      = flag.Bool("init", false, "Generate stub methods for converters")
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

	units, err := generator.Decide(i, *flagInit, info.Name, *flagOutputDir)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	for _, unit := range units {
		err := unit.Generate()
		if err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}
	}
	fmt.Println("All files successfully generated")
}

func findInterface(file *types.File, ifaceName string) *types.Interface {
	for i := range file.Interfaces {
		if file.Interfaces[i].Name == ifaceName {
			return &file.Interfaces[i]
		}
	}
	return nil
}
