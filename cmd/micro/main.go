package main

import (
	"bytes"
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"strings"

	"github.com/devimteam/microgen/pkg/microgen"

	"github.com/devimteam/microgen/gen"
	"github.com/devimteam/microgen/internal/bootstrap"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

var (
	flagConfig = flag.String("config", "microgen.yaml", "path to configuration file")
)

func main() {
	if len(os.Args) == 1 {
		fmt.Fprintln(os.Stderr, "interface name is required. Example: '$ microgen UserService'")
		os.Exit(1)
	}
	ifaceArg := os.Args[len(os.Args)-1]
	pkgs, err := parser.ParseDir(token.NewFileSet(), ".", nonTestFilter, 0)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	var iface microgen.Interface
	var pkgName string
	for _, pkg := range pkgs {
		// Remove all unexported declarations
		if !ast.PackageExports(pkg) {
			continue
		}
		if ast.FilterPackage(pkg, func(name string) bool { return name == ifaceArg }) {
			pkgName = pkg.Name
			var (
				emptyInterface bool
			)
			ast.Inspect(pkg, func(node ast.Node) bool {
				switch n := node.(type) {
				case *ast.GenDecl:
					return true
				case *ast.TypeSpec:
					if n.Name == nil || n.Name.Name != ifaceArg {
						return false
					}
					i, ok := n.Type.(*ast.InterfaceType)
					if !ok {
						return false
					}
					iface.Docs = parseComments(n.Doc)
					if i.Methods == nil {
						emptyInterface = true
						return false
					}
					for _, method := range i.Methods.List {
						if method == nil {
							continue
						}

					}
				default:
					return false
				}
				return false
			})
			if emptyInterface {
				fmt.Fprintln(os.Stderr, "interface has no methods")
				os.Exit(1)
			}
			break // found
		}
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
	err = bootstrap.Run(trim(cfg.Plugins), ifaceArg, currentPkg, pkgName)
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

func parseComments(group *ast.CommentGroup) (comments []string) {
	if group == nil {
		return
	}
	for _, comment := range group.List {
		comments = append(comments, comment.Text)
	}
	return
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
	Plugins []string
}
