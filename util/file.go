package util

import (
	"fmt"
	astparser "go/parser"
	"go/token"
	"os"
	"path/filepath"

	"github.com/vetcher/godecl"
	"github.com/vetcher/godecl/types"
)

func ParseFile(filename string) (*types.File, error) {
	currentDir, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	path := filepath.Join(currentDir, filename)
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
