package util

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/vetcher/godecl"
	"github.com/vetcher/godecl/types"
)

func ParseFile(filename string) (*types.File, error) {
	return godecl.ParseFile(filename)
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
