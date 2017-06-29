package util

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
)

type FuncArg struct {
	Name string
	Type string
}

type Func struct {
	Args []*FuncArg
}

func GetInterfaceFuncTypes(path, iface string) ([]*ast.Field, error) {
	fset := token.NewFileSet()
	fileAST, err := parser.ParseFile(fset, path, nil, 0)
	if err != nil {
		return nil, err
	}

	for _, decl := range fileAST.Decls {
		decl, ok := decl.(*ast.GenDecl)
		if !ok || decl.Tok != token.TYPE {
			continue
		}
		for _, spec := range decl.Specs {
			spec := spec.(*ast.TypeSpec)
			if spec.Name.Name != iface {
				continue
			}

			ifaceTypeSpec, ok := spec.Type.(*ast.InterfaceType)
			if !ok {
				continue
			}

			return ifaceTypeSpec.Methods.List, nil
		}
	}

	return nil, fmt.Errorf("type %s not found in %s", iface, path)
}
