package generator

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
)

const MkdirPermissions = 0777

type newFileStrategy struct {
	outputDir string
}

func (s newFileStrategy) Write(f Renderer, t Template) error {
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

	err = s.Save(f, outpath)
	if err != nil {
		return fmt.Errorf("error when save file: %v", err)
	}
	fmt.Println(filepath.Join(s.outputDir, t.Path()))
	return nil
}

// Copied from original github.com/dave/jennifer/jen.go func Save()
func (s newFileStrategy) Save(f Renderer, filename string) error {
	buf := &bytes.Buffer{}
	if err := f.Render(buf); err != nil {
		return err
	}
	if err := ioutil.WriteFile(filename, buf.Bytes(), 0644); err != nil {
		return err
	}
	return nil
}

func NewFileStrategy(dir string) Strategy {
	return newFileStrategy{
		outputDir: dir,
	}
}
