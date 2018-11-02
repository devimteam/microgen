package microgen

import (
	"context"
	"fmt"
	"reflect"
)

func ValidateInterface(iface *Interface) error {
	var errs []error
	if len(iface.Methods) == 0 {
		errs = append(errs, fmt.Errorf("%s does not have any methods", iface.Name))
	}
	for _, m := range iface.Methods {
		errs = append(errs, validateFunction(m)...)
	}
	return composeErrors(errs...)
}

// Rules:
// * First argument is context.Context.
// * Last result is error.
// * All params have names.
// * Not a raw interface, function or struct.
func validateFunction(fn Method) (errs []error) {
	// do not validate methods with `//microgen:-`
	if FetchTags(fn.Docs, "//"+Microgen).Has("-") {
		return nil
	}
	if !isContextFirst(fn.Args) {
		errs = append(errs, fmt.Errorf("%s: first argument should be of type context.Context", fn.Name))
	}
	if !isErrorLast(fn.Results) {
		errs = append(errs, fmt.Errorf("%s: last result should be of type error", fn.Name))
	}
	for _, param := range append(fn.Args, fn.Results...) {
		if param.Name == "" {
			errs = append(errs, fmt.Errorf("%s: unnamed parameter of type %s", fn.Name, param.Type.String()))
		}
		if isRawInterface(param.Type) {
			errs = append(errs, fmt.Errorf("%s: non empty interface '%s %s' is not allowed, delcare it outside", fn.Name, param.Name, param.Type.String()))
		}
		if isRawStruct(param.Type) {
			errs = append(errs, fmt.Errorf("%s: raw struct '%s %s' is not allowed, declare it outside", fn.Name, param.Name, param.Type.String()))
		}
		if isRawFunc(param.Type) {
			errs = append(errs, fmt.Errorf("%s: raw function '%s %s' is not allowed, declare it outside", fn.Name, param.Name, param.Type.String()))
		}
	}
	return errs
}

func isRawInterface(t reflect.Type) bool {
	if t.PkgPath() != "" {
		return false
	}
	switch t.Kind() {
	case reflect.Ptr, reflect.Slice, reflect.Array, reflect.Chan:
		return isRawInterface(t.Elem())
	case reflect.Map:
		return isRawInterface(t.Key()) && isRawInterface(t.Elem())
	case reflect.Interface:
		if t.NumMethod() == 0 {
			return true
		}
		return false
	default:
		return false
	}
}

func isRawStruct(t reflect.Type) bool {
	if t.PkgPath() != "" {
		return false
	}
	switch t.Kind() {
	case reflect.Ptr, reflect.Slice, reflect.Array, reflect.Chan:
		return isRawStruct(t.Elem())
	case reflect.Map:
		return isRawStruct(t.Key()) && isRawStruct(t.Elem())
	case reflect.Struct:
		return true
	default:
		return false
	}
}

func isRawFunc(t reflect.Type) bool {
	if t.PkgPath() != "" {
		return false
	}
	switch t.Kind() {
	case reflect.Ptr, reflect.Slice, reflect.Array, reflect.Chan:
		return isRawFunc(t.Elem())
	case reflect.Map:
		return isRawFunc(t.Key()) && isRawFunc(t.Elem())
	case reflect.Func:
		return true
	default:
		return false
	}
}

var (
	contextType = reflect.TypeOf(new(context.Context)).Elem()
	errorType   = reflect.TypeOf(new(error)).Elem()
)

func isContextFirst(fields []Var) bool {
	return len(fields) != 0 && fields[0].Type == contextType
}

func isErrorLast(fields []Var) bool {
	n := len(fields)
	return n != 0 && fields[n-1].Type == errorType
}
