package parser

import (
	"errors"
	"fmt"
	"go/ast"
	"go/token"
	"path"
	"strings"
)

var (
	ErrCouldNotResolvePackage = errors.New("could not resolve package")
	ErrUnexpectedFieldType    = errors.New("provided fields have unexpected type")
)

type Interface struct {
	PackageName    string
	Docs           []string
	Name           string
	Imports        []*Import
	FuncSignatures []*FuncSignature
}

type FuncSignature struct {
	Name    string
	Params  []*FuncField
	Results []*FuncField
}

type Import struct {
	Alias string
	Path  string
}

// Basic field struct.
// Used for tiny parameters and results representation.
type FuncField struct {
	Name      string
	Package   *Import
	Type      string
	IsPointer bool
	IsArray   bool
}

func resolveImportByAlias(imports []*Import, alias string) (*Import, error) {
	// try to find by alias
	for _, imp := range imports {
		if imp.Alias == alias {
			return imp, nil
		}
	}

	// try to find by last package path
	for _, imp := range imports {
		_, pname := path.Split(imp.Path)
		if alias == pname {
			return imp, nil
		}
	}

	return nil, ErrCouldNotResolvePackage
}

// Build list of function signatures by provided
// AST of file and interface name.
func ParseInterface(f *ast.File, ifaceName string) (*Interface, error) {
	imports, err := parseImports(f)
	if err != nil {
		return nil, fmt.Errorf("error when fetch imports: %v", err)
	}

	genDecl, typeSpec, err := findTypeByName(f, ifaceName)
	if err != nil {
		return nil, fmt.Errorf("could not find type: %v", err)
	}

	ifaceSpec, ok := typeSpec.Type.(*ast.InterfaceType)
	if !ok {
		return nil, ErrUnexpectedFieldType
	}

	funcSignatures, err := parseFuncSignatures(ifaceSpec.Methods.List, imports)
	if err != nil {
		return nil, fmt.Errorf("error when parse func signatures: %v", funcSignatures)
	}

	return &Interface{
		Imports:        imports,
		PackageName:    f.Name.Name,
		Name:           ifaceName,
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

func parseImports(f *ast.File) ([]*Import, error) {
	for _, decl := range f.Decls {
		decl, ok := decl.(*ast.GenDecl)
		if !ok || decl.Tok != token.IMPORT {
			continue
		}

		var imports []*Import

		for _, spec := range decl.Specs {
			spec, ok := spec.(*ast.ImportSpec)
			if !ok {
				return nil, ErrUnexpectedFieldType
			}

			imp := Import{}

			imp.Path = strings.Trim(spec.Path.Value, `"`)

			if spec.Name != nil {
				imp.Alias = spec.Name.Name
			} else {
				_, imp.Alias = path.Split(imp.Path)
			}

			imports = append(imports, &imp)
		}

		return imports, nil
	}

	return nil, nil
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
func parseFuncSignatures(fields []*ast.Field, imports []*Import) ([]*FuncSignature, error) {
	var funcs []*FuncSignature
	for _, field := range fields {
		funcType, ok := field.Type.(*ast.FuncType)
		if !ok {
			return nil, ErrUnexpectedFieldType
		}

		f := &FuncSignature{
			// TODO: resolve magic number
			Name: field.Names[0].Name,
		}

		for _, param := range funcType.Params.List {
			for _, paramName := range param.Names {
				ff := &FuncField{Name: paramName.Name}
				err := parseField(ff, param.Type, imports)
				if err != nil {
					return nil, fmt.Errorf("could not parse field %s: %v", ff.Name, err)
				}
				f.Params = append(f.Params, ff)
			}
		}

		for _, result := range funcType.Results.List {
			for _, resultName := range result.Names {
				ff := &FuncField{Name: resultName.Name}
				err := parseField(ff, result.Type, imports)
				if err != nil {
					return nil, fmt.Errorf("could not parse field %s: %v", ff.Name, err)
				}
				f.Results = append(f.Results, ff)
			}
		}

		funcs = append(funcs, f)
	}

	return funcs, nil
}

func parseField(ff *FuncField, flType interface{}, imports []*Import) error {
	switch t := flType.(type) {
	case *ast.Ident:
		ff.Type = t.Name
	case *ast.SelectorExpr:
		var err error
		ff.Type = t.Sel.Name
		ff.Package, err = resolveImportByAlias(imports, t.X.(*ast.Ident).Name)
		if err != nil {
			return err
		}
	case *ast.StarExpr:
		ff.IsPointer = true
		err := parseField(ff, t.X, imports)
		if err != nil {
			return err
		}
	case *ast.ArrayType:
		ff.IsArray = true
		err := parseField(ff, t.Elt, imports)
		if err != nil {
			return err
		}
	}

	return nil
}
