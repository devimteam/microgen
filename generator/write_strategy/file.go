package write_strategy

import (
	"bytes"
	"fmt"
	"go/format"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
)

const (
	MkdirPermissions = 0777

	// This hack needs for normal code formatting.
	// It makes formatter think, that declared code is top-level.
	// Without this hack formatter adds separators (tabs) to beginning of every line.
	formatTrick = "package T\n"

	NewFileMark    = "New"
	AppendFileMark = "Add"
	IgnoreFileMark = "Ign"
)

type createFileStrategy struct {
	absPath string
	relPath string
}

func (s createFileStrategy) Write(renderer Renderer) error {
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
	fmt.Println(NewFileMark, filepath.Join(s.absPath, s.relPath))
	return nil
}

// Copied from original github.com/dave/jennifer/jen.go func Save()
func (s createFileStrategy) Save(f Renderer, filename string) error {
	buf := &bytes.Buffer{}
	if err := f.Render(buf); err != nil {
		return err
	}
	formatted, err := format.Source(buf.Bytes())
	if err != nil {
		return fmt.Errorf("error when format source: %v", err)
	}
	if err := ioutil.WriteFile(filename, formatted, 0644); err != nil {
		return err
	}
	return nil
}

func NewCreateFileStrategy(absPath, relPath string) Strategy {
	return createFileStrategy{
		absPath: absPath,
		relPath: relPath,
	}
}

type appendFileStrategy struct {
	absPath string
	relPath string
}

func NewAppendToFileStrategy(absPath, relPath string) Strategy {
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
	fmt.Println(AppendFileMark, filepath.Join(s.absPath, s.relPath))
	return nil
}

func (s appendFileStrategy) Save(renderer Renderer, filename string) error {
	buf := &bytes.Buffer{}
	if err := renderer.Render(buf); err != nil {
		return err
	}
	// Use trick for top-level formatting.
	formatted, err := format.Source(append([]byte(formatTrick), buf.Bytes()...))
	if err != nil {
		return fmt.Errorf("error when format source: %v", err)
	}

	f, err := os.OpenFile(filename, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	if _, err = f.Write(formatted[len(formatTrick):]); err != nil {
		return err
	}
	return nil
}
