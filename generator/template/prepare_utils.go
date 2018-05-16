package template

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/vetcher/go-astra"
	"github.com/vetcher/go-astra/types"
)

func ParseFile(filename string) (*types.File, error) {
	return astra.ParseFile(filename)
}

var parsedCache = map[string]*types.File{}

func parsePackage(path string) (*types.File, error) {
	path = filepath.Dir(path)
	if file, ok := parsedCache[path]; ok {
		return file, nil
	}
	files, err := astra.ParsePackage(path, astra.AllowAnyImportAliases)
	if err != nil {
		return nil, err
	}
	file, err := astra.MergeFiles(files)
	if err != nil {
		return nil, err
	}
	parsedCache[path] = file
	return file, nil
}

func statFile(absPath, relPath string) error {
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
