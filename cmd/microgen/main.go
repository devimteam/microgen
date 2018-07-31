package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/devimteam/microgen/generator"
	mstrings "github.com/devimteam/microgen/generator/strings"
	"github.com/devimteam/microgen/generator/template"
	lg "github.com/devimteam/microgen/logger"
	"github.com/vetcher/go-astra"
	"github.com/vetcher/go-astra/types"
)

const (
	Version = generator.Version
)

var (
	flagFileName     = flag.String("file", "service.go", "Path to input file with interface.")
	flagOutputDir    = flag.String("out", ".", "Output directory.")
	flagHelp         = flag.Bool("help", false, "Show help.")
	flagVerbose      = flag.Int("v", 1, "Sets microgen verbose level.")
	flagDebug        = flag.Bool("debug", false, "Print all microgen messages. Equivalent to -v=100.")
	flagGenProtofile = flag.String(".proto", "", "Package field in protobuf file. If not empty, service.proto file will be generated.")
	flagGenMain      = flag.Bool(generator.MainTag, false, "Generate main.go file.")
)

func init() {
	flag.Parse()
}

func main() {
	lg.Logger.Level = *flagVerbose
	if *flagDebug {
		lg.Logger.Level = 100
	}
	lg.Logger.Logln(1, "@microgen", Version)
	if *flagHelp || *flagFileName == "" {
		flag.Usage()
		os.Exit(0)
	}

	lg.Logger.Logln(4, "Source file:", *flagFileName)
	info, err := astra.ParseFile(*flagFileName)
	if err != nil {
		lg.Logger.Logln(0, "fatal:", err)
		os.Exit(1)
	}

	i := findInterface(info)
	if i == nil {
		lg.Logger.Logln(0, "fatal: could not find interface with @microgen tag")
		lg.Logger.Logln(4, "All founded interfaces:")
		lg.Logger.Logln(4, listInterfaces(info.Interfaces))
		os.Exit(1)
	}

	if err := generator.ValidateInterface(i); err != nil {
		lg.Logger.Logln(0, "validation:", err)
		os.Exit(1)
	}

	ctx, err := prepareContext(*flagFileName, i)
	if err != nil {
		lg.Logger.Logln(0, "fatal:", err)
		os.Exit(1)
	}

	absOutputDir, err := filepath.Abs(*flagOutputDir)
	if err != nil {
		lg.Logger.Logln(0, "fatal:", err)
		os.Exit(1)
	}
	units, err := generator.ListTemplatesForGen(ctx, i, absOutputDir, *flagFileName, *flagGenProtofile, *flagGenMain)
	if err != nil {
		lg.Logger.Logln(0, "fatal:", err)
		os.Exit(1)
	}
	for _, unit := range units {
		err := unit.Generate(ctx)
		if err != nil && err != generator.EmptyStrategyError {
			lg.Logger.Logln(0, "fatal:", unit.Path(), err)
			os.Exit(1)
		}
	}
	lg.Logger.Logln(1, "all files successfully generated")
}

func listInterfaces(ii []types.Interface) string {
	var s string
	for _, i := range ii {
		s = s + fmt.Sprintf("\t%s(%d methods, %d embedded interfaces)\n", i.Name, len(i.Methods), len(i.Interfaces))
	}
	return s
}

func prepareContext(filename string, iface *types.Interface) (context.Context, error) {
	ctx := context.Background()
	p, err := astra.ResolvePackagePath(filename)
	if err != nil {
		return nil, err
	}
	ctx = template.WithSourcePackageImport(ctx, p)

	set := template.TagsSet{}
	genTags := mstrings.FetchTags(iface.Docs, generator.TagMark+generator.MicrogenMainTag)
	for _, tag := range genTags {
		set.Add(tag)
	}
	ctx = template.WithTags(ctx, set)
	return ctx, nil
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
