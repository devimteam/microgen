package generator

import (
	"fmt"

	"github.com/devimteam/microgen/generator/template"
	"github.com/devimteam/microgen/util"
	"github.com/vetcher/go-astra/types"
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
	for _, param := range append(fn.Args, fn.Results...) {
		if param.Name == "" {
			errs = append(errs, fmt.Errorf("%s: unnamed parameter of type %s", fn.Name, param.Type.String()))
		}
		if iface := types.TypeInterface(param.Type); iface != nil && !iface.(types.TInterface).Interface.IsEmpty() {
			errs = append(errs, fmt.Errorf("%s: non empty interface %s is not allowed, delcare it outside", fn.Name, param.String()))
		}
		if strct := types.TypeStruct(param.Type); strct != nil {
			errs = append(errs, fmt.Errorf("%s: raw struct %s is not allowed, declare it outside", fn.Name, param.Name))
		}
		if f := types.TypeFunction(param.Type); f != nil {
			errs = append(errs, fmt.Errorf("%s: raw function %s is not allowed, declare it outside", fn.Name, param.Name))
		}
	}
	if template.FetchHttpMethodTag(fn.Docs) == "GET" && !isArgumentsAllowSmartPath(fn) {
		errs = append(errs, fmt.Errorf("%s: can't use GET method with provided arguments", fn.Name))
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
