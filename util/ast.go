package util

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
)

type FuncField struct {
	Name string
	Type string
}

type Func struct {
	Name string
	Params []*FuncField
	Results []*FuncField
}

func ParseInterface(path, ifaceName string) ([]*Func, error) {
	f, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
	if err != nil {
		return nil, fmt.Errorf("unable to parse file: %v", err)
	}

	typeSpec, err := findType(f, ifaceName)

	ifaceSpec, ok := typeSpec.Type.(*ast.InterfaceType)
	if !ok {
		return nil, fmt.Errorf("type '%s' is not interface", ifaceName)
	}

	return parseInterfaceFieldList(ifaceSpec.Methods.List)
}

func findType(f *ast.File, name string) (*ast.TypeSpec, error) {
	for _, decl := range f.Decls {
		decl, ok := decl.(*ast.GenDecl)
		if !ok || decl.Tok != token.TYPE {
			continue
		}
		for _, spec := range decl.Specs {
			spec := spec.(*ast.TypeSpec)
			if spec.Name.Name != name {
				continue
			}

			return spec, nil
		}
	}

	return nil, fmt.Errorf("type '%s' not found in %s", name, f.Name.Name)
}

func parseInterfaceFieldList(fields []*ast.Field) ([]*Func, error) {
	var funcs []*Func
	for _, field := range fields {
		funcType := field.Type.(*ast.FuncType)

		f := &Func{
			// TODO: resolve magic number
			Name: field.Names[0].Name,
		}

		for _, param := range funcType.Params.List {
			for _, paramName := range param.Names {
				paramType := param.Type.(*ast.Ident).Name
				f.Params = append(f.Params, &FuncField{
					Name: paramName.Name,
					Type: paramType,
				})
			}
		}

		for _, result := range funcType.Results.List {
			for _, resultName := range result.Names {
				resultType := result.Type.(*ast.Ident).Name
				f.Results = append(f.Results, &FuncField{
					Name: resultName.Name,
					Type: resultType,
				})
			}
		}

		funcs = append(funcs, f)
	}

	return funcs, nil
}