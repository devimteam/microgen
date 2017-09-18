package generator

import (
	"fmt"

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

func validateFunction(fn *types.Function) (errs []error) {
	for _, param := range fn.Args {
		if param.Name == "" {
			errs = append(errs, fmt.Errorf("%s: unnamed argument of type %s", fn.Name, param.Type.String()))
		}
		if param.Type.IsInterface {
			errs = append(errs, fmt.Errorf("%s: argument error: raw interface (%s) type is not allowed, declare it as type", fn.Name, param.Type.String()))
		}
	}
	for _, param := range fn.Results {
		if param.Name == "" {
			errs = append(errs, fmt.Errorf("%s: unnamed result of type %s", fn.Name, param.Type.String()))
		}
		if param.Type.IsInterface {
			errs = append(errs, fmt.Errorf("%s: result error: raw interface (%s) type is not allowed, declare it as type", fn.Name, param.Type.String()))
		}
	}
	return
}
