package write_method

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
	absPath string
	relPath string
}

func (s newFileStrategy) Write(renderer Renderer) error {
	outpath, err := filepath.Abs(filepath.Join(s.absPath, s.relPath))
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

	err = s.Save(renderer, outpath)
	if err != nil {
		return fmt.Errorf("error when save file: %v", err)
	}
	fmt.Println(filepath.Join(s.absPath, s.relPath))
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

func NewFileMethod(absPath, relPath string) Method {
	return newFileStrategy{
		absPath: absPath,
		relPath: relPath,
	}
}

type appendFileStrategy struct {
	absPath string
	relPath string
}

func AppendToFileStrategy(absPath, relPath string) Method {
	return appendFileStrategy{
		absPath: absPath,
		relPath: relPath,
	}
}

func (s appendFileStrategy) Write(renderer Renderer) error {
	outpath, err := filepath.Abs(filepath.Join(s.absPath, s.relPath))
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

	if _, err = os.Stat(outpath); os.IsNotExist(err) {
		f, err := os.Create(outpath)
		if err != nil {
			return fmt.Errorf("can't create %s: error: %v", outpath, err)
		}
		f.Close()
	}

	err = s.Save(renderer, outpath)
	if err != nil {
		return fmt.Errorf("error when save file: %v", err)
	}
	fmt.Println(filepath.Join(s.absPath, s.relPath))
	return nil
}

func (s appendFileStrategy) Save(renderer Renderer, filename string) error {
	buf := &bytes.Buffer{}
	if err := renderer.Render(buf); err != nil {
		return err
	}
	f, err := os.OpenFile(filename, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	if _, err = f.Write(buf.Bytes()); err != nil {
		return err
	}
	return nil
}
