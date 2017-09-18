package template

import (
	"github.com/devimteam/microgen/util"
	"github.com/vetcher/godecl/types"
	. "github.com/vetcher/jennifer/jen"
)

const (
	PackagePathGoKitEndpoint      = "github.com/go-kit/kit/endpoint"
	PackagePathContext            = "context"
	PackagePathGoKitLog           = "github.com/go-kit/kit/log"
	PackagePathTime               = "time"
	PackagePathGoogleGRPC         = "google.golang.org/grpc"
	PackagePathNetContext         = "golang.org/x/net/context"
	PackagePathGoKitTransportGRPC = "github.com/go-kit/kit/transport/grpc"
)

func structFieldName(field *types.Variable) *Statement {
	return Id(util.ToUpperFirst(field.Name))
}

// Remove from function fields context if it is first in slice
func removeContextIfFirst(fields []types.Variable) []types.Variable {
	if len(fields) > 0 && fields[0].Type.Import != nil && fields[0].Type.Import.Package == PackagePathContext {
		return fields[1:]
	}
	return fields
}

// Renders struct field.
//
//  	Visit *entity.Visit `json:"visit"`
//
func structField(field *types.Variable) *Statement {
	s := structFieldName(field)
	s.Add(fieldType(&field.Type))
	s.Tag(map[string]string{"json": util.ToSnakeCase(field.Name)})
	return s
}

// Renders func params for definition.
//
//  	visit *entity.Visit, err error
//
func funcDefinitionParams(fields []types.Variable) *Statement {
	c := &Statement{}
	c.ListFunc(func(g *Group) {
		for _, field := range fields {
			g.Id(util.ToLowerFirst(field.Name)).Add(fieldType(&field.Type))
		}
	})
	return c
}

// Renders field type for given func field.
//
//  	*repository.Visit
//
func fieldType(field *types.Type) *Statement {
	c := &Statement{}
	if field.IsArray {
		c.Index()
	}

	if field.IsPointer {
		c.Op("*")
	}
	if field.IsMap {
		m := field.Map()
		return c.Map(fieldType(&m.Key)).Add(fieldType(&m.Value))
	}
	if field.IsInterface {
		c.Interface()
	}
	if field.Import != nil {
		c.Qual(field.Import.Package, field.Name)
	} else {
		c.Id(field.Name)
	}

	return c
}

// Renders key/value pairs wrapped in Dict for provided fields.
//
//		Err:    err,
//		Result: result,
//
func dictByVariables(fields []types.Variable) Dict {
	return DictFunc(func(d Dict) {
		for _, field := range fields {
			d[structFieldName(&field)] = Id(util.ToLowerFirst(field.Name))
		}
	})
}

// Render list of function receivers by signature.Result.
//
//		Ans1, ans2, AnS3 -> ans1, ans2, anS3
//
func paramNames(fields []types.Variable) *Statement {
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
func methodDefinition(obj string, signature *types.Function) *Statement {
	return Func().
		Params(Id(util.FirstLowerChar(obj)).Op("*").Id(obj)).
		Id(signature.Name).
		Params(funcDefinitionParams(signature.Args)).
		Params(funcDefinitionParams(signature.Results))
}

// TODO: Resolve this hardcoded path
func protobufPath(svcName string) string {
	return "gitlab.devim.team/protobuf/" + svcName
}
