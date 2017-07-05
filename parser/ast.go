package parser

import (
	"fmt"
	"go/ast"
	"go/token"
)

type Interface struct {
	PackageName    string
	Docs           []string
	FuncSignatures []*FuncSignature
}

type FuncSignature struct {
	Name string
	Params []*FuncField
	Results []*FuncField
}

// Basic field struct.
// Used for tiny parameters and results representation.
type FuncField struct {
	Name string
	Type string
}


// Build list of function signatures by provided
// AST of file and interface name.
func ParseInterface(f *ast.File, ifaceName string) (*Interface, error) {
	genDecl, typeSpec, err := findTypeByName(f, ifaceName)
	if err != nil {
		return nil, fmt.Errorf("could not find type: %v", err)
	}

	ifaceSpec, ok := typeSpec.Type.(*ast.InterfaceType)
	if !ok {
		return nil, fmt.Errorf("type '%s' is not interface", ifaceName)
	}

	funcSignatures, err := parseFuncSignatures(ifaceSpec.Methods.List)
	if err != nil {
		return nil, err
	}

	return &Interface{
		PackageName:    getPackageName(f),
		Docs:           parseDocs(genDecl),
		FuncSignatures: funcSignatures,
	}, nil
}

// Returns type spec by name from provided AST of file.
func findTypeByName(f *ast.File, name string) (*ast.GenDecl, *ast.TypeSpec, error) {
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

			return decl, spec, nil
		}
	}

	return nil, nil, fmt.Errorf("type '%s' not found in %s", name, f.Name.Name)
}

func getPackageName(f *ast.File) string {
	return f.Name.Name
}

// Parse doc of interface generic declaration.
func parseDocs(d *ast.GenDecl) []string {
	if d.Doc == nil {
		return nil
	}

	var res []string
	for _, c := range d.Doc.List {
		res = append(res, c.Text)
	}
	return res
}

// Returns function signature by provided method list.
// Method list represents as array of pointers to ast.Field.
func parseFuncSignatures(fields []*ast.Field) ([]*FuncSignature, error) {
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