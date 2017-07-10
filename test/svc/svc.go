package svc

import (
	c "context"
	"go/ast"
	"os"
)

// This is an interface of the service.
// Yay
type StringService interface {
	Uppercase(ctx *c.Context, in, in2 string, in3 int) (cntx c.Context, err error)
	Lowercase(ctx c.Context, in *os.File) (out *ast.File, err error)
}
