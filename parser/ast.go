package parser

import (
	"fmt"
	"go/ast"
	"go/token"
)

// Basic field struct.
// Used for tiny parameters and results representation.
type FuncField struct {
	Name string
	Type string
}

// Func signature representation.
type FuncSignature struct {
	Name string
	Params []*FuncField
	Results []*FuncField
}

// Build list of function signatures by provided
// AST of file and interface name.
func GetInterfaceFuncSignatures(f *ast.File, ifaceName string) ([]*FuncSignature, error) {
	typeSpec, err := getTypeSpecByName(f, ifaceName)
	if err != nil {
		return nil, fmt.Errorf("could not find type: %v", err)
	}

	ifaceSpec, ok := typeSpec.Type.(*ast.InterfaceType)
	if !ok {
		return nil, fmt.Errorf("type '%s' is not interface", ifaceName)
	}

	return parseInterfaceFieldList(ifaceSpec.Methods.List)
}

// Returns type spec by name from provided AST of file.
func getTypeSpecByName(f *ast.File, name string) (*ast.TypeSpec, error) {
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

// Returns function signature by provided method list.
// Method list represents as array of pointers to ast.Field.
func parseInterfaceFieldList(fields []*ast.Field) ([]*FuncSignature, error) {
	var funcs []*FuncSignature
	for _, field := range fields {
		funcType, ok := field.Type.(*ast.FuncType)
		if !ok {
			return nil, fmt.Errorf("provided fields not implement ast.FuncType")
		}

		f := &FuncSignature{
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