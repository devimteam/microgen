package main

import (
	"flag"
	"os"
	"strings"
	"sync"

	"github.com/devimteam/microgen/generator"
	lg "github.com/devimteam/microgen/logger"
	"github.com/devimteam/microgen/util"
	"github.com/vetcher/godecl/types"
)

const (
	Version = generator.Version
)

var (
	flagFileName  = flag.String("file", "service.go", "Name of file where described interface definition")
	flagOutputDir = flag.String("out", ".", "Output directory")
	flagHelp      = flag.Bool("help", false, "Show help")
	flagForce     = flag.Bool("force", false, "Overwrite all files, as it generates for the first time")
	flagVerbose   = flag.Int("v", 1, "Verbose log level")
)

func init() {
	flag.Parse()
}

func main() {
	lg.Logger.Level = *flagVerbose
	lg.Logger.Logln(1, "@microgen", Version)
	if *flagHelp || *flagFileName == "" {
		flag.Usage()
		os.Exit(0)
	}

	info, err := util.ParseFile(*flagFileName)
	if err != nil {
		lg.Logger.Logln(0, "fatal:", err)
		os.Exit(1)
	}

	i := findInterface(info)
	if i == nil {
		lg.Logger.Logln(0, "fatal: could not find interface with @microgen tag")
		os.Exit(1)
	}

	if err := generator.ValidateInterface(i); err != nil {
		lg.Logger.Logln(0, "validation:", err)
		os.Exit(1)
	}

	units, err := generator.ListTemplatesForGen(i, *flagForce, info.Name, *flagOutputDir, *flagFileName)
	if err != nil {
		lg.Logger.Logln(0, "fatal:", err)
		os.Exit(1)
	}
	var wg sync.WaitGroup
	wg.Add(len(units))
	for _, x := range units {
		unit := x
		go func() {
			defer wg.Done()
			err := unit.Generate()
			if err != nil && err != generator.EmptyStrategyError {
				lg.Logger.Logln(0, "fatal:", err)
				os.Exit(1)
			}
		}()
	}
	wg.Wait()
	lg.Logger.Logln(1, "all files successfully generated")
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
		if strings.HasPrefix(str, generator.TagMark+generator.MicrogenMainTag) {
			return true
		}
	}
	return false
}
