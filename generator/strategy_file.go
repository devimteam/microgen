package generator

import (
	"fmt"
	"os"
	"path"
	"path/filepath"

	"github.com/dave/jennifer/jen"
)

const MkdirPermissions = 0777

type fileStrategy struct {
	outputDir string
}

func (s fileStrategy) Write(f *jen.File, t Template) error {
	outpath, err := filepath.Abs(filepath.Join(s.outputDir, t.Path()))
	if err != nil {
		return fmt.Errorf("unable to resolve path: %v", err)
	}
	dir := path.Dir(outpath)

	_, err = os.Stat(dir)

	if os.IsNotExist(err) {
		err = os.MkdirAll(dir, MkdirPermissions)
		if err != nil {
			return fmt.Errorf("unable to create directory %s: %v", outpath, err)
		}
	} else if err != nil {
		return fmt.Errorf("could not stat file: %v", err)
	}

	err = f.Save(outpath)
	if err != nil {
		return fmt.Errorf("error when save file: %v", err)
	}
	fmt.Println(filepath.Join(s.outputDir, t.Path()))
	return nil
}

func NewFileStrategy(dir string) Strategy {
	return fileStrategy{
		outputDir: dir,
	}
}
