package template

import (
	. "github.com/dave/jennifer/jen"
	"github.com/devimteam/microgen/parser"
	"github.com/devimteam/microgen/util"
)

const (
	PackagePathGoKitEndpoint = "github.com/go-kit/kit/endpoint"
	PackagePathContext       = "context"
	PackagePathGoKitLog      = "github.com/go-kit/kit/log"
	PackagePathTime          = "time"
)

func structFieldName(field *parser.FuncField) *Statement {
	return Id(util.ToUpperFirst(field.Name))
}

// Check if function field type of context.Context
func checkFieldIsContext(field *parser.FuncField) bool {
	if field.Package != nil && field.Package.Path == PackagePathContext {
		return true
	}
	return false
}

// Renders struct field.
//
//  	Visit *entity.Visit `json:"visit"`
//
func structField(field *parser.FuncField) *Statement {
	s := structFieldName(field)
	s.Add(fieldType(field))
	s.Tag(map[string]string{"json": util.ToSnakeCase(field.Name)})
	return s
}

// Renders func params for definition.
//
//  	visit *entity.Visit, err error
//
func funcDefinitionParams(fields []*parser.FuncField) *Statement {
	c := &Statement{}
	c.ListFunc(func(g *Group) {
		for _, field := range fields {
			g.Id(util.ToLowerFirst(field.Name)).Add(fieldType(field))
		}
	})
	return c
}

// Renders field type for given func field.
//
//  	*repository.Visit
//
func fieldType(field *parser.FuncField) *Statement {
	c := &Statement{}

	if field.IsArray {
		c.Index()
	}

	if field.IsPointer {
		c.Op("*")
	}

	if field.Package != nil {
		c.Qual(field.Package.Path, field.Type)
	} else {
		c.Id(field.Type)
	}

	return c
}

// Renders key/value pairs wrapped in Dict for provided fields.
//
//		Err:    err,
//		Result: result,
//
func dictByFuncFields(fields []*parser.FuncField) Dict {
	return DictFunc(func(d Dict) {
		for i, field := range fields {
			if i == 0 && checkFieldIsContext(field) {
				continue
			}
			d[structFieldName(field)] = Id(util.ToLowerFirst(field.Name))
		}
	})
}

// Renders func params for function call.
//
//  	req.Visit, req.Err
//
func funcCallParams(obj string, fields []*parser.FuncField) *Statement {
	var list []Code
	for _, field := range fields {
		list = append(list, Id(obj).Dot(util.ToUpperFirst(field.Name)))
	}
	return List(list...)
}

// Render list of function receivers by signature.Result.
//
//		Ans1, ans2, AnS3 -> ans1, ans2, anS3
//
func funcReceivers(fields []*parser.FuncField) *Statement {
	var list []Code
	for _, field := range fields {
		list = append(list, Id(util.ToLowerFirst(field.Name)))
	}
	return List(list...)
}

// Render method call with receivers and params.
//
//		count := svc.Count(ctx, req.Text, req.Symbol)
//
func serviceMethodCallWithReceivers(service, request string, signature *parser.FuncSignature) *Statement {
	return funcReceivers(signature.Results).Op(":=").Id(service).Dot(signature.Name).Call(funcCallParams(request, signature.Params))
}

// Render full method definition with receiver, method name, args and results.
//
//		func (e *Endpoints) Count(ctx context.Context, text string, symbol string) (count int)
//
func methodDefinition(obj string, signature *parser.FuncSignature) *Statement {
	return Func().
		Params(Id(util.FirstLowerChar(obj)).Op("*").Id(obj)).
		Id(signature.Name).
		Params(funcDefinitionParams(signature.Params)).
		Params(funcDefinitionParams(signature.Results))
}
