package main

import (
	"bytes"
	"flag"
	"fmt"
	"go/parser"
	"go/token"
	"os"
	"strings"

	"github.com/devimteam/microgen/internal/pkgpath"

	"github.com/devimteam/microgen/internal/bootstrap"
	lg "github.com/devimteam/microgen/logger"
	toml "github.com/pelletier/go-toml"
	"github.com/pkg/errors"
)

var (
	flagVerbose = flag.Int("v", lg.Common, "Sets microgen verbose level.")
	flagDebug   = flag.Bool("debug", false, "Print all microgen messages. Equivalent to -v=100.")
	flagConfig  = flag.String("config", "microgen.yaml", "path to configuration file")
	flagDry     = flag.Bool("dry", false, "Do everything except writing files.")
	flagKeep    = flag.Bool("keep", false, "Keeps bootstrapped file after execution.")
	flagForce   = flag.Bool("force", false, "Forcing microgen to overwrite files, that was marked as 'edited manually'")
)

func init() {
	flag.Parse()
}

func main() {
	if len(os.Args) == 1 {
		fmt.Fprintln(os.Stderr, "interface name is required. Example: '$ microgen UserService'")
		os.Exit(1)
	}
	if *flagVerbose < lg.Critical {
		*flagVerbose = lg.Critical
	}
	lg.Logger.Level = *flagVerbose
	if *flagDebug {
		lg.Logger.Level = lg.Debug
		lg.Logger.Logln(lg.Debug, "Debug logs mode in on")
	}
	ifaceArg := os.Args[len(os.Args)-1]
	pkgs, err := parser.ParseDir(token.NewFileSet(), ".", nonTestFilter, parser.ParseComments)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	iface, err := getInterface(ifaceArg, pkgs)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	cfg, err := processConfig(*flagConfig)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	currentPkg, err := pkgpath.GetPkgPath(".", true)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	err = bootstrap.Run(trim(expandEnv(cfg.Import)), iface, currentPkg, *flagKeep)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Println("Done")
}

func nonTestFilter(info os.FileInfo) bool {
	return !strings.HasSuffix(info.Name(), "_test.go")
}

func trim(ss []string) []string {
	for i := range ss {
		ss[i] = strings.Trim(ss[i], `"`)
	}
	return ss
}

func expandEnv(ss []string) []string {
	for i := range ss {
		ss[i] = os.ExpandEnv(ss[i])
	}
	return ss
}

func processConfig(pathToConfig string) (*config, error) {
	file, err := os.Open(pathToConfig)
	if err != nil {
		return nil, errors.WithMessage(err, "open file")
	}
	var rawToml bytes.Buffer
	_, err = rawToml.ReadFrom(file)
	if err != nil {
		return nil, errors.WithMessage(err, "read from config")
	}
	var cfg config
	err = toml.NewDecoder(&rawToml).Decode(&cfg)
	if err != nil {
		return nil, errors.WithMessage(err, "unmarshal config")
	}
	return &cfg, nil
}

type config struct {
	Import []string `toml:"import"`
}
