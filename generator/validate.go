package generator

import (
	"fmt"

	mstrings "github.com/devimteam/microgen/generator/strings"
	"github.com/devimteam/microgen/generator/template"
	"github.com/devimteam/microgen/util"
	"github.com/vetcher/godecl/types"
)

func ValidateInterface(iface *types.Interface) error {
	var errs []error
	for _, m := range iface.Methods {
		errs = append(errs, validateFunction(m)...)
	}
	return util.ComposeErrors(errs...)
}

// Rules:
// * First argument is context.Context.
// * Last result is error.
// * All params have names.
func validateFunction(fn *types.Function) (errs []error) {
	if !template.IsContextFirst(fn.Args) {
		errs = append(errs, fmt.Errorf("%s: first argument should be of type context.Context", fn.Name))
	}
	if !template.IsErrorLast(fn.Results) {
		errs = append(errs, fmt.Errorf("%s: last result should be of type error", fn.Name))
	}
	for _, param := range fn.Args {
		if param.Name == "" {
			errs = append(errs, fmt.Errorf("%s: unnamed argument of type %s", fn.Name, param.Type.String()))
		}
	}
	for _, param := range fn.Results {
		if param.Name == "" {
			errs = append(errs, fmt.Errorf("%s: unnamed result of type %s", fn.Name, param.Type.String()))
		}
	}
	if mstrings.ContainTag(fn.Docs, TagMark+HttpSmartPathTag) && template.FetchHttpMethodTag(fn.Docs) != "GET" {
		errs = append(errs, fmt.Errorf("%s: %s should be used only with GET method", fn.Name, HttpSmartPathTag))
	}
	if mstrings.ContainTag(fn.Docs, TagMark+HttpSmartPathTag) && !isArgumentsAllowSmartPath(fn) {
		errs = append(errs, fmt.Errorf("%s: can't use %s with provided arguments", fn.Name, HttpSmartPathTag))
	}
	return
}

func isArgumentsAllowSmartPath(fn *types.Function) bool {
	for _, arg := range template.RemoveContextIfFirst(fn.Args) {
		if !canInsertToPath(&arg) {
			return false
		}
	}
	return true
}

var insertableToUrlTypes = []string{"string", "int", "int32", "int64", "uint", "uint32", "uint64"}

// We can make url variable from string, int, int32, int64, uint, uint32, uint64
func canInsertToPath(p *types.Variable) bool {
	name := types.TypeName(p.Type)
	return name != nil && util.IsInStringSlice(*name, insertableToUrlTypes)
}
