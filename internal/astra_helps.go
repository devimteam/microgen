package internal

import (
	"strconv"
	"strings"

	"github.com/dave/jennifer/jen"
	mstrings "github.com/devimteam/microgen/generator/strings"
	"github.com/devimteam/microgen/pkg/microgen"
	"github.com/vetcher/go-astra/types"
)

const (
	PackagePathContext = "context"
)

// Remove from function fields context if it is first in slice
func RemoveContextIfFirst(fields []types.Variable) []types.Variable {
	if IsContextFirst(fields) {
		return fields[1:]
	}
	return fields
}

func IsContextFirst(fields []types.Variable) bool {
	if len(fields) == 0 {
		return false
	}
	name := types.TypeName(fields[0].Type)
	return name != nil &&
		types.TypeImport(fields[0].Type) != nil &&
		types.TypeImport(fields[0].Type).Package == PackagePathContext &&
		*name == "Context"
}

// Remove from function fields error if it is last in slice
func RemoveErrorIfLast(fields []types.Variable) []types.Variable {
	if IsErrorLast(fields) {
		return fields[:len(fields)-1]
	}
	return fields
}

func IsErrorLast(fields []types.Variable) bool {
	if len(fields) == 0 {
		return false
	}
	name := types.TypeName(fields[len(fields)-1].Type)
	return name != nil &&
		types.TypeImport(fields[len(fields)-1].Type) == nil &&
		*name == "error"
}

func Var(ctx microgen.Context, v types.Variable) *jen.Statement {
	return jen.Id(v.Name).Add(VarType(ctx, v.Type, true))
}

func VarType(ctx microgen.Context, field types.Type, allowEllipsis bool) *jen.Statement {
	c := &jen.Statement{}
	imported := false
	for field != nil {
		switch f := field.(type) {
		case types.TImport:
			if f.Import != nil {
				c.Qual(f.Import.Package, "")
				imported = true
			}
			field = f.Next
		case types.TName:
			if !imported && !types.IsBuiltin(f) {
				c.Qual(ctx.SourcePackageImport, f.TypeName)
			} else {
				c.Id(f.TypeName)
			}
			field = nil
		case types.TArray:
			if f.IsSlice {
				c.Index()
			} else if f.ArrayLen > 0 {
				c.Index(jen.Lit(f.ArrayLen))
			}
			field = f.Next
		case types.TMap:
			return c.Map(VarType(ctx, f.Key, false)).Add(VarType(ctx, f.Value, false))
		case types.TPointer:
			c.Op(strings.Repeat("*", f.NumberOfPointers))
			field = f.Next
		case types.TInterface:
			mhds := interfaceType(ctx, f.Interface)
			return c.Interface(mhds...)
		case types.TEllipsis:
			if allowEllipsis {
				c.Op("...")
			} else {
				c.Index()
			}
			field = f.Next
		case types.TChan:
			if f.Direction == types.ChanDirRecv {
				c.Op("<-")
			}
			c.Chan()
			if f.Direction == types.ChanDirSend {
				c.Op("<-")
			}
			field = f.Next
		default:
			c.Id(f.String())
			field = nil
		}
	}
	return c
}

// Render full method definition with receiver, method name, args and results.
//
//		func Count(ctx context.Context, text string, symbol string) (count int)
//
func functionDefinition(ctx microgen.Context, signature *types.Function) *jen.Statement {
	return jen.Id(signature.Name).
		Params(funcDefinitionParams(ctx, signature.Args)).
		Params(funcDefinitionParams(ctx, signature.Results))
}

func interfaceType(ctx microgen.Context, p *types.Interface) (code []jen.Code) {
	for _, x := range p.Methods {
		code = append(code, functionDefinition(ctx, x))
	}
	return
}

// Renders func params for definition.
//
//  	visit *entity.Visit, err error
//
func funcDefinitionParams(ctx microgen.Context, fields []types.Variable) *jen.Statement {
	c := &jen.Statement{}
	c.ListFunc(func(g *jen.Group) {
		for _, field := range fields {
			g.Id(mstrings.ToLowerFirst(field.Name)).Add(VarType(ctx, field.Type, true))
		}
	})
	return c
}

// Render full method definition with receiver, method name, args and results.
//
//		func (e Endpoints) Count(ctx context.Context, text string, symbol string) (count int)
//
func MethodDefinition(ctx microgen.Context, obj string, signature *types.Function) *jen.Statement {
	return jen.Func().
		Params(jen.Id(Rec(obj)).Id(obj)).
		Add(functionDefinition(ctx, signature))
}

func Rec(name string) string {
	return mstrings.LastUpperOrFirst(name)
}

// Render list of function receivers by signature.Result.
//
//		Ans1, ans2, AnS3 -> ans1, ans2, anS3
//
func ParamNames(fields []types.Variable) *jen.Statement {
	var list []jen.Code
	for _, field := range fields {
		v := jen.Id(mstrings.ToLowerFirst(field.Name))
		if types.IsEllipsis(field.Type) {
			v.Op("...")
		}
		list = append(list, v)
	}
	return jen.List(list...)
}

type NormalizedFunction struct {
	types.Function
	Parent *types.Function
}

const (
	normalArgPrefix    = "arg_"
	normalResultPrefix = "res_"
)

func NormalizeFunction(signature *types.Function) *NormalizedFunction {
	newFunc := &NormalizedFunction{Parent: signature}
	newFunc.Name = signature.Name
	newFunc.Args = NormalizeVariables(signature.Args, normalArgPrefix)
	newFunc.Results = NormalizeVariables(signature.Results, normalResultPrefix)
	return newFunc
}

func NormalizeVariables(old []types.Variable, prefix string) (new []types.Variable) {
	for i := range old {
		v := old[i]
		v.Name = prefix + strconv.Itoa(i)
		new = append(new, v)
	}
	return
}

// Return name of error, if error is last result, else return `err`
func NameOfLastResultError(fn *types.Function) string {
	if IsErrorLast(fn.Results) {
		return fn.Results[len(fn.Results)-1].Name
	}
	return "err"
}

func DictByNormalVariables(fields []types.Variable, normals []types.Variable) jen.Dict {
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
func FirstArgName(signature *types.Function) string {
	return mstrings.ToLowerFirst(signature.Args[0].Name)
}
