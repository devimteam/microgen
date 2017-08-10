package template

import (
	. "github.com/dave/jennifer/jen"
	"github.com/devimteam/microgen/parser"
	"github.com/devimteam/microgen/util"
)

const (
	PackagePathGoKitEndpoint  = "github.com/go-kit/kit/endpoint"
	PackagePathContext        = "context"
	PackagePathGoKitLog       = "github.com/go-kit/kit/log"
	PackagePathTime           = "time"
	PackagePathTransportLayer = "github.com/devimteam/go-kit/transportlayer"
)

func structFieldName(field *parser.FuncField) *Statement {
	return Id(util.ToUpperFirst(field.Name))
}

// Remove from function fields context if it is first in slice
func removeContextIfFirst(fields []*parser.FuncField) []*parser.FuncField {
	if len(fields) > 0 && fields[0].Package != nil && fields[0].Package.Path == PackagePathContext {
		return fields[1:]
	}
	return fields
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
		for _, field := range fields {
			d[structFieldName(field)] = Id(util.ToLowerFirst(field.Name))
		}
	})
}

// Render list of function receivers by signature.Result.
//
//		Ans1, ans2, AnS3 -> ans1, ans2, anS3
//
func paramNames(fields []*parser.FuncField) *Statement {
	var list []Code
	for _, field := range fields {
		list = append(list, Id(util.ToLowerFirst(field.Name)))
	}
	return List(list...)
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
