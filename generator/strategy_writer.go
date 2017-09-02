package generator

import (
	"fmt"
	"io"
)

type writerStrategy struct {
	writer io.Writer
}

func (s writerStrategy) Write(f Renderer, t Template) error {
	err := f.Render(s.writer)
	if err != nil {
		return fmt.Errorf("render error: %v", err)
	}
	return nil
}

func WriterStrategy(writer io.Writer) Strategy {
	return writerStrategy{
		writer: writer,
	}
}
