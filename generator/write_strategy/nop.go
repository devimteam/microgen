package write_strategy

import (
	"fmt"
	"path/filepath"
)

type nopStrategy struct {
	absPath string
	relPath string
}

// Do nothing strategy
func NewNopStrategy(absPath, relPath string) Strategy {
	return nopStrategy{
		absPath: absPath,
		relPath: relPath,
	}
}

func (s nopStrategy) Write(Renderer) error {
	fmt.Println(IgnoreFileMark, filepath.Join(s.absPath, s.relPath))
	return nil
}

func (s nopStrategy) Save(Renderer, string) error {
	return nil
}
