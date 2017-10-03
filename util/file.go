package util

import (
	"fmt"
	astparser "go/parser"
	"go/token"
	"path/filepath"

	"os"

	"github.com/vetcher/godecl"
	"github.com/vetcher/godecl/types"
)

func ParseFile(filename string) (*types.File, error) {
	path, err := filepath.Abs(filename)
	if err != nil {
		return nil, fmt.Errorf("can not filepath.Abs: %v", err)
	}
	fset := token.NewFileSet()
	tree, err := astparser.ParseFile(fset, path, nil, astparser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("error when parse file: %v\n", err)
	}
	info, err := godecl.ParseFile(tree)
	if err != nil {
		return nil, fmt.Errorf("error when parsing info from file: %v\n", err)
	}
	return info, nil
}

func StatFile(absPath, relPath string) error {
	outpath, err := filepath.Abs(filepath.Join(absPath, relPath))
	if err != nil {
		return fmt.Errorf("unable to resolve path: %v", err)
	}

	fileInfo, err := os.Stat(outpath)
	if os.IsNotExist(err) || os.IsPermission(err) {
		return err
	}
	if fileInfo.IsDir() {
		return fmt.Errorf("%s is dir", outpath)
	}
	return nil
}

func FindFunctionByName(fns []types.Function, name string) *types.Function {
	for i := range fns {
		if fns[i].Name == name {
			return &fns[i]
		}
	}
	return nil
}
