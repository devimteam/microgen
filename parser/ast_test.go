package parser

import (
	"testing"
	"go/parser"
	"go/token"
)

var src = `
package svc

type StringService interface {
	Uppercase(in, in2 string, in3 int) (out string, err error)
}
`

func TestGetInterfaceFuncSignatures(t *testing.T) {
	f, err := parser.ParseFile(token.NewFileSet(), "", src, 0)
	if err != nil {
		t.Errorf("unable to parse file: %v", err)
	}

	fs, err := GetInterfaceFuncSignatures(f, "StringService")
	if err != nil {
		t.Errorf("could not get interface func signatures: %v", err)
	}

	if fs[0].Name != "Uppercase" {
		t.Errorf("invalid parsing of interface")
	}
}