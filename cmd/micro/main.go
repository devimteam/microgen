package main

import (
	"bytes"
	"flag"
	"fmt"
	"go/parser"
	"go/token"
	"os"
	"strings"

	"github.com/devimteam/microgen/gen"
	"github.com/devimteam/microgen/internal"
	"github.com/devimteam/microgen/internal/bootstrap"
	"github.com/devimteam/microgen/logger"
	toml "github.com/pelletier/go-toml"
	"github.com/pkg/errors"
)

var (
	flagConfig  = flag.String("config", "microgen.yaml", "path to configuration file")
	flagVerbose = flag.Int("v", logger.Common, "Sets microgen verbose level.")
	flagDebug   = flag.Bool("debug", false, "Print all microgen messages. Equivalent to -v=100.")
)

func init() {
	flag.Parse()
}

func main() {
	if len(os.Args) == 1 {
		fmt.Fprintln(os.Stderr, "interface name is required. Example: '$ microgen UserService'")
		os.Exit(1)
	}
	ifaceArg := os.Args[len(os.Args)-1]
	pkgs, err := parser.ParseDir(token.NewFileSet(), ".", nonTestFilter, parser.ParseComments)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	iface, err := internal.GetInterface(ifaceArg, pkgs)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	cfg, err := processConfig(*flagConfig)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	currentPkg, err := gen.GetPkgPath(".", true)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	err = bootstrap.Run(trim(expandEnv(cfg.Import)), iface, currentPkg)
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
