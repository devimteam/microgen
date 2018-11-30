package main

import (
	"fmt"
	"go/ast"
	"os"

	"github.com/devimteam/microgen/pkg/microgen"
	"github.com/pkg/errors"
)

func getInterface(name string, pkgs map[string]*ast.Package) (iface microgen.Interface, err error) {
	found := false
	for _, pkg := range pkgs {
		// Remove all unexported declarations
		if !ast.PackageExports(pkg) {
			continue
		}
		ast.Inspect(pkg, func(node ast.Node) bool {
			switch n := node.(type) {
			case *ast.Package:
				return true
			case *ast.File:
				return true
			case *ast.GenDecl:
				for i := range n.Specs {
					t, ok := n.Specs[i].(*ast.TypeSpec)
					if !ok {
						continue
					}
					if t.Name == nil || t.Name.Name != name {
						continue
					}
					i, ok := t.Type.(*ast.InterfaceType)
					if !ok {
						continue
					}
					found = true
					if i.Methods == nil {
						fmt.Fprintln(os.Stderr, "interface has no methods")
						os.Exit(1)
					}
					iface.Name = name
					iface.Docs = parseComments(n.Doc, t.Doc, t.Comment)
					for _, method := range i.Methods.List {
						if method == nil {
							continue
						}
						fn, ok := method.Type.(*ast.FuncType)
						if !ok {
							fmt.Fprintln(os.Stderr, "embedding other interfaces is not supported")
							os.Exit(1)
						}
						iface.Methods = append(iface.Methods, microgen.Method{
							Docs:    parseComments(method.Doc),
							Name:    safeName(method.Names),
							Args:    namesToVars(namesFromFieldList(fn.Params)),
							Results: namesToVars(namesFromFieldList(fn.Results)),
						})
					}
				}
				return false // stop inspecting
			default:
				return false
			}
			return false
		})
		if found {
			iface.PackageName = pkg.Name
			break
		}
	}
	if !found {
		err = errors.New("interface not found")
	}
	return iface, err
}

func safeName(names []*ast.Ident) string {
	if len(names) == 0 {
		return ""
	}
	if names[0] == nil {
		return ""
	}
	return names[0].Name
}

func safeNames(names []*ast.Ident) []string {
	res := make([]string, len(names))
	for i := range names {
		if names[i] == nil {
			continue
		}
		res[i] = names[i].Name
	}
	return res
}

func namesFromFieldList(list *ast.FieldList) []string {
	if list == nil {
		return nil
	}
	var res []string
	for i := range list.List {
		if list.List[i] == nil {
			continue
		}
		res = append(res, safeNames(list.List[i].Names)...)
	}
	return res
}

func namesToVars(ss []string) []microgen.Var {
	x := make([]microgen.Var, len(ss))
	for i := range ss {
		x[i].Name = ss[i]
	}
	return x
}

func parseComments(groups ...*ast.CommentGroup) (comments []string) {
	for _, group := range groups {
		if group == nil {
			return
		}
		for _, comment := range group.List {
			comments = append(comments, comment.Text)
		}
	}
	return
}
