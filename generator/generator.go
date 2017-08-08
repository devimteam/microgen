package generator

import (
	"fmt"
	"io"
	"path/filepath"

	"github.com/devimteam/microgen/parser"
)

const CodeHeader = `This file was automatically generated by "microgen" utility.
Please, do not edit.`

type Generator interface {
	Generate() error
}

type generator struct {
	templates []Template
	iface     *parser.Interface
	outputDir string
}

func NewGenerator(ts []Template, iface *parser.Interface, outputDir string) Generator {
	return &generator{
		templates: ts,
		iface:     iface,
		outputDir: outputDir,
	}
}

func (g *generator) Generate() error {
	for _, t := range g.templates {
		c := t.Render(g.iface)
		c.PackageComment(CodeHeader)

		path, err := filepath.Abs(filepath.Join(g.outputDir, t.Path()))
		if err != nil {
			return fmt.Errorf("abs path error: %v", err)
		}

		err = c.Save(path)
		if err != nil {
			return fmt.Errorf("error when save file: %v", err)
		}
	}

	return nil
}

type writerGenerator struct {
	templates []Template
	iface     *parser.Interface
	writer    io.Writer
}

func NewWriterGenerator(ts []Template, iface *parser.Interface, writer io.Writer) Generator {
	return &writerGenerator{
		templates: ts,
		iface:     iface,
		writer:    writer,
	}
}

func (g *writerGenerator) Generate() error {
	for _, t := range g.templates {
		c := t.Render(g.iface)
		c.PackageComment(CodeHeader)
		err := c.Render(g.writer)
		if err != nil {
			return fmt.Errorf("render error: %v", err)
		}
	}
	return nil
}
