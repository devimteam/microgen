package generator

import (
	"fmt"

	"github.com/devimteam/microgen/generator/template"
	"github.com/devimteam/microgen/util"
	"github.com/vetcher/godecl/types"
)

func ValidateInterface(iface *types.Interface) error {
	var errs []error
	for _, m := range iface.Methods {
		errs = append(errs, validateFunction(m)...)
	}
	return util.ComposeErrors(errs)
}

func isInterface(p *types.Type) bool {
	if p.IsInterface {
		return true
	}
	if p.IsMap {
		return isInterface(&p.Map.Key) || isInterface(&p.Map.Value)
	}
	return false
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
	return
}
