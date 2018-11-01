package internal

import (
	"context"
	"reflect"
	"strconv"

	"github.com/dave/jennifer/jen"
	mstrings "github.com/devimteam/microgen/generator/strings"
	"github.com/devimteam/microgen/pkg/microgen"
)

const (
	PackagePathContext = "context"
)

var (
	ContextType = reflect.TypeOf(new(context.Context)).Elem()
	ErrorType   = reflect.TypeOf(new(error)).Elem()
)

// Remove from function fields context if it is first in slice
func RemoveContextIfFirst(fields []microgen.Var) []microgen.Var {
	if IsContextFirst(fields) {
		return fields[1:]
	}
	return fields
}

func IsContextFirst(fields []microgen.Var) bool {
	return len(fields) != 0 && fields[0].Type == ContextType
}

// Remove from function fields error if it is last in slice
func RemoveErrorIfLast(fields []microgen.Var) []microgen.Var {
	if IsErrorLast(fields) {
		return fields[:len(fields)-1]
	}
	return fields
}

func IsErrorLast(fields []microgen.Var) bool {
	n := len(fields)
	return n != 0 && fields[n-1].Type == ErrorType
}

func Var(v microgen.Var) *jen.Statement {
	return jen.Id(v.Name).Add(VarType(v.Type, true))
}

func VarType(field reflect.Type, allowEllipsis bool) *jen.Statement {
	c := &jen.Statement{}
	for field != nil {
		if field.PkgPath() != "" {
			c.Qual(field.PkgPath(), field.Name())
			break
		}
		switch field.Kind() {
		case reflect.Array:
			c.Index(jen.Lit(field.Len()))
			field = field.Elem()
		case reflect.Chan:
			switch field.ChanDir() {
			case reflect.RecvDir:
				c.Op("<-").Chan()
			case reflect.SendDir:
				c.Chan().Op("<-")
			default:
				c.Chan()
			}
			field = field.Elem()
		case reflect.Func:
			field = nil
		case reflect.Interface:
			if field.NumMethod() == 0 {
				c.Interface()
			} else if field == ErrorType {
				c.Error()
			}
			field = nil
		case reflect.Map:
			c.Map(VarType(field.Key(), false)).Add(VarType(field.Elem(), false))
			field = nil
		case reflect.Ptr:
			c.Op("*")
			field = field.Elem()
		case reflect.Slice:
			c.Index()
			field = field.Elem()
			field.Name()
		default:
			c.Id(field.String())
			field = nil
		}
	}
	return c
}

// Render full method definition with receiver, method name, args and results.
//
//		func Count(ctx context.Context, text string, symbol string) (count int)
//
func functionDefinition(signature microgen.Method) *jen.Statement {
	return jen.Id(signature.Name).
		Params(funcDefinitionParams(signature.Args)).
		Params(funcDefinitionParams(signature.Results))
}

func interfaceType(p *microgen.Interface) (code []jen.Code) {
	for _, x := range p.Methods {
		code = append(code, functionDefinition(x))
	}
	return
}

// Renders func params for definition.
//
//  	visit *entity.Visit, err error
//
func funcDefinitionParams(fields []microgen.Var) *jen.Statement {
	c := &jen.Statement{}
	c.ListFunc(func(g *jen.Group) {
		for _, field := range fields {
			g.Id(mstrings.ToLowerFirst(field.Name)).Add(VarType(field.Type, true))
		}
	})
	return c
}

// Render full method definition with receiver, method name, args and results.
//
//		func (e Endpoints) Count(ctx context.Context, text string, symbol string) (count int)
//
func MethodDefinition(obj string, signature microgen.Method) *jen.Statement {
	return jen.Func().
		Params(jen.Id(Rec(obj)).Id(obj)).
		Add(functionDefinition(signature))
}

func Rec(name string) string {
	return mstrings.LastUpperOrFirst(name)
}

// Render list of function receivers by signature.Result.
//
//		Ans1, ans2, AnS3 -> ans1, ans2, anS3
//
func ParamNames(fields []microgen.Var) *jen.Statement {
	var list []jen.Code
	for _, field := range fields {
		v := jen.Id(mstrings.ToLowerFirst(field.Name))
		list = append(list, v)
	}
	return jen.List(list...)
}

type NormalizedFunction struct {
	microgen.Method
	Parent microgen.Method
}

const (
	normalArgPrefix    = "arg_"
	normalResultPrefix = "res_"
)

func NormalizeFunction(signature microgen.Method) *NormalizedFunction {
	newFunc := &NormalizedFunction{Parent: signature}
	newFunc.Name = signature.Name
	newFunc.Args = NormalizeVariables(signature.Args, normalArgPrefix)
	newFunc.Results = NormalizeVariables(signature.Results, normalResultPrefix)
	return newFunc
}

func NormalizeVariables(old []microgen.Var, prefix string) (new []microgen.Var) {
	for i := range old {
		v := old[i]
		v.Name = prefix + strconv.Itoa(i)
		new = append(new, v)
	}
	return
}

// Return name of error, if error is last result, else return `err`
func NameOfLastResultError(fn microgen.Method) string {
	if IsErrorLast(fn.Results) {
		return fn.Results[len(fn.Results)-1].Name
	}
	return "err"
}

func DictByNormalVariables(fields []microgen.Var, normals []microgen.Var) jen.Dict {
	if len(fields) != len(normals) {
		panic("len of fields and normals not the same")
	}
	return jen.DictFunc(func(d jen.Dict) {
		for i, field := range fields {
			d[jen.Id(mstrings.ToUpperFirst(field.Name))] = jen.Id(mstrings.ToLowerFirst(normals[i].Name))
		}
	})
}

// For custom ctx in service interface (e.g. context or ctxxx).
func FirstArgName(signature microgen.Method) string {
	return mstrings.ToLowerFirst(signature.Args[0].Name)
}
