package internal

import (
	"strings"

	"github.com/devimteam/microgen/gen"
	mstrings "github.com/devimteam/microgen/generator/strings"
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

func FuncFields(fields []types.Variable, pkg string, ellipsis bool) []interface{} {
	res := make([]interface{}, len(fields))
	for i, field := range fields {
		res[i] = gen.Dot(mstrings.ToLowerFirst(field.Name), " ", varType(fields[i].Type, pkg, ellipsis), ",")
	}
	return res
}

func varType(t types.Type, pkg string, ellipsis bool) (res []interface{}) {
	imported := false
	for t != nil {
		switch f := t.(type) {
		case types.TImport:
			if f.Import != nil {
				res = append(res, gen.Imp(f.Import.Package), ".")
				imported = true
			}
			t = f.Next
		case types.TName:
			if !imported && !types.IsBuiltin(t) {
				res = append(res, gen.Imp(pkg), ".")
			}
			res = append(res, f.TypeName)
		case types.TArray:
			if f.IsSlice {
				res = append(res, "[]")
			} else {
				res = append(res, "[", f.ArrayLen, "]")
			}
		case types.TMap:
			return append(res, "map[", varType(f.Key, pkg, false), "]", varType(f.Value, pkg, false))
		case types.TPointer:
			res = append(res, strings.Repeat("*", f.NumberOfPointers))
			t = f.Next
		case types.TInterface:
			return append(res, "interface{", interfaceMethods(f.Interface, pkg, true), "}")
		case types.TEllipsis:
			if ellipsis {
				res = append(res, "...")
			} else {
				res = append(res, "[]")
			}
			t = f.Next
		case types.TChan:
			if f.Direction == types.ChanDirRecv {
				res = append(res, "<-")
			}
			res = append(res, "chan")
			if f.Direction == types.ChanDirSend {
				res = append(res, "<-")
			}
			t = f.Next
		default:
			res = append(res, f.String())
			t = nil
		}
	}
	return res
}

func interfaceMethods(iface *types.Interface, pkg string, ellipsis bool) []interface{} {
	lm := len(iface.Methods)
	res := make([]interface{}, lm+len(iface.Interfaces))
	for i, m := range iface.Methods {
		res[i] = append(FuncDefinition(m, pkg), '\n')
	}
	for i, emb := range iface.Interfaces {
		res[i+lm] = varType(emb.Type, pkg, ellipsis)
	}
	return res
}

func FuncDefinition(fn *types.Function, pkg string) []interface{} {
	return gen.Dot(fn.Name, "(", FuncFields(fn.Args, pkg, true), ") (",
		FuncFields(fn.Args, pkg, false), ")")
}
