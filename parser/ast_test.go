package parser

import (
	"go/parser"
	"go/token"
	"testing"
)

var src = `
package svc

import (
	"context"
	"github.com/fakeuser/testpackage"
)

type StringService interface {
	Uppercase(ctx context.Context, in, in2 string, in3 int, yay testpackage.TestType) (out string, err error)
}
`

func TestGetInterfaceFuncSignatures(t *testing.T) {
	//t.Skip()
	f, err := parser.ParseFile(token.NewFileSet(), "", src, 0)
	if err != nil {
		t.Errorf("unable to parse file: %v", err)
	}

	fs, err := ParseInterface(f, "StringService")
	if err != nil {
		t.Errorf("could not get interface func signatures: %v", err)
	}

	if fs.FuncSignatures[0].Name != "Uppercase" {
		t.Errorf("invalid parsing of interface")
	}

	if fs.FuncSignatures[0].Params[4].Name != "yay" {
		t.Errorf("invalid parsing of parameters")
	}
}
