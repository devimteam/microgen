package templates_test

import (
	goparser "go/parser"
	"go/token"
	"testing"

	"bytes"

	"github.com/devimteam/microgen/generator"
	parser "github.com/devimteam/microgen/parser"
	"io/ioutil"
	"github.com/devimteam/microgen/generator/template"
)

func TestExchangeForCountSvc(t *testing.T) {
	src, err := ioutil.ReadFile("service.txt")
	if err != nil {
		t.Fatalf("read source file error: %v", err)
	}
	out, err := ioutil.ReadFile("exchange.txt")
	if err != nil {
		t.Fatalf("read out file error: %v", err)
	}

	f, err := goparser.ParseFile(token.NewFileSet(), "", src, 0)
	if err != nil {
		t.Errorf("unable to parse file: %v", err)
	}
	fs, err := parser.ParseInterface(f, "StringService")
	if err != nil {
		t.Errorf("could not get interface func signatures: %v", err)
	}
	buf := bytes.NewBuffer([]byte{})
	gen := generator.NewGenerator([]generator.Template{
		&template.ExchangeTemplate{},
	}, fs, generator.NewWriterStrategy(buf))
	err = gen.Generate()
	if err != nil {
		t.Errorf("unable to generate: %v", err)
	}
	if bytes.Equal(buf.Bytes(), out) {
		t.Errorf("Got:\n\n%v", buf.String())
	}
}
