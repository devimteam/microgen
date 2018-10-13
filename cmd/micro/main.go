package main

import (
	"bytes"
	"fmt"
	"os"
	"strings"

	"github.com/devimteam/microgen/gen"
	"github.com/devimteam/microgen/internal/bootstrap"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

func main() {
	// todo: check package for containing each of interfaces
	/*
		pkgs, err := parser.ParseDir(token.NewFileSet(), ".", nonTestFilter, 0)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		for _, pkg := range pkgs {
			// Remove all unexported declarations
			if !ast.PackageExports(pkg) {
				continue
			}
		}*/
	cfg, err := processConfig("micro.yaml")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	currentPkg, err := gen.GetPkgPath(".", true)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	err = bootstrap.Run(trim(cfg.Plugins), cfg.interfaces(), currentPkg)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Println("DONE")
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
	err = yaml.NewDecoder(&rawToml).Decode(&cfg)
	if err != nil {
		return nil, errors.WithMessage(err, "unmarshal config")
	}
	return &cfg, nil
}

type config struct {
	Plugins    []string
	Interfaces []yaml.MapItem
}

func (c config) interfaces() []string {
	var ss []string
	for i := range c.Interfaces {
		if s, ok := c.Interfaces[i].Key.(string); ok {
			ss = append(ss, s)
		}
	}
	return ss
}
