package generator

import (
	"fmt"
	"io"

	"github.com/dave/jennifer/jen"
)

type writerStrategy struct {
	writer io.Writer
}

func (s writerStrategy) Write(f *jen.File, t Template) error {
	err := f.Render(s.writer)
	if err != nil {
		return fmt.Errorf("render error: %v", err)
	}
	return nil
}

func NewWriterStrategy(writer io.Writer) Strategy {
	return writerStrategy{
		writer: writer,
	}
}
