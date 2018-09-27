package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/devimteam/microgen/generator"
	"github.com/devimteam/microgen/internal"
	lg "github.com/devimteam/microgen/logger"
	"github.com/vetcher/go-astra"
	"github.com/vetcher/go-astra/types"
)

const (
	Version = generator.Version
)

var (
	flagSource       = flag.String("src", ".", "Path to input package with interfaces.")
	flagDstDir       = flag.String("dst", ".", "Destiny path.")
	flagHelp         = flag.Bool("help", false, "Show help.")
	flagVerbose      = flag.Int("v", 1, "Sets microgen verbose level.")
	flagDebug        = flag.Bool("debug", false, "Print all microgen messages. Equivalent to -v=100.")
	flagGenProtofile = flag.String(".proto", "", "Package field in protobuf file. If not empty, service.proto file will be generated.")
	flagGenMain      = flag.Bool(generator.MainTag, false, "Generate main.go file.")
	flagInterface    = flag.String("interface", "", "Name of the target interface. If package contains one interface with microgen tags, arg may be omitted.")
)

func init() {
	if !flag.Parsed() {
		flag.Parse()
	}
}

func main() {
	begin := time.Now()
	defer func() { lg.Logger.Logln(1, "Full:", time.Since(begin)) }()
	if *flagVerbose < 0 {
		*flagVerbose = 0
	}
	lg.Logger.Level = *flagVerbose
	if *flagDebug {
		lg.Logger.Level = 100
	}
	lg.Logger.Logln(1, "microgen", Version)
	if *flagHelp || *flagSource == "" {
		flag.Usage()
		os.Exit(0)
	}

	lg.Logger.Logln(4, "Source package:", *flagSource)
	info, err := astra.GetPackage(*flagSource)
	if err != nil {
		lg.Logger.Logln(0, "fatal:", err)
		os.Exit(1)
	}

	ii := findInterfaces(info)
	iface, err := selectInterface(ii)
	if err != nil {
		lg.Logger.Logln(0, "fatal:", err)
		lg.Logger.Logln(4, "All founded interfaces:")
		lg.Logger.Logln(4, listInterfaces(info.Interfaces))
		os.Exit(1)
	}

	ctx, err := prepareContext(*flagSource, *flagDstDir, iface)
	if err != nil {
		lg.Logger.Logln(0, "fatal:", err)
		os.Exit(1)
	}

	absOutputDir, err := filepath.Abs(*flagDstDir)
	if err != nil {
		lg.Logger.Logln(0, "fatal:", err)
		os.Exit(1)
	}
	units, err := generator.ListTemplatesForGen(ctx, iface, absOutputDir, *flagGenProtofile, *flagGenMain)
	if err != nil {
		lg.Logger.Logln(0, "fatal:", err)
		os.Exit(1)
	}
	lg.Logger.Logln(3, "Preparing:", time.Since(begin))
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

func prepareContext(sourceFileName, dstPath string, iface *types.Interface) (context.Context, error) {
	ctx := context.Background()
	p, err := internal.ResolvePackagePath(sourceFileName)
	if err != nil {
		return nil, err
	}
	ctx = internal.WithSourcePackageImport(ctx, p)
	dp, err := internal.ResolvePackagePath(dstPath)
	if err != nil {
		return nil, err
	}
	ctx = internal.WithDstPkgImport(ctx, dp)
	ctx = internal.WithSource(ctx, sourceFileName)
	ctx = internal.WithDst(ctx, dstPath)

	genTags := internal.FetchTags(iface.Docs, generator.TagMark+generator.MicrogenMainTag)
	ctx = internal.WithTags(ctx, genTags)

	return ctx, nil
}

func findInterfaces(file *types.File) []*types.Interface {
	var ifaces []*types.Interface
	for i := range file.Interfaces {
		if docsContainMicrogenTag(file.Interfaces[i].Docs) {
			ifaces = append(ifaces, &file.Interfaces[i])
		}
	}
	return ifaces
}

func selectInterface(ii []*types.Interface) (*types.Interface, error) {
	if len(ii) == 0 {
		return ii[0], nil
	}
	if *flagInterface == "" {
		return nil, fmt.Errorf("%d interfaces founded, but -interface is empty. Please, provide an interface name", len(ii))
	}
	for i := range ii {
		if ii[i].Name == *flagInterface {
			return ii[i], nil
		}
	}
	return nil, fmt.Errorf("%s interface not found, but %d others are available", *flagInterface, len(ii))
}

func docsContainMicrogenTag(strs []string) bool {
	for _, str := range strs {
		if strings.HasPrefix(str, generator.TagMark+generator.MicrogenMainTag) {
			return true
		}
	}
	return false
}
