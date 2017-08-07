package template

import (
	. "github.com/dave/jennifer/jen"
	"github.com/devimteam/microgen/parser"
	"github.com/devimteam/microgen/util"
)

const PackageAliasGoKit = "github.com/go-kit/kit/endpoint"

type EndpointsTemplate struct {
}

// Renders endpoints file.
//
//  package visitsvc
//
//  import (
//  	"context"
//
//  	"github.com/go-kit/kit/endpoint"
//  	"gitlab.devim.team/microservices/visitsvc/entity"
//  )
//
//  type Endpoints struct {
//  	CreateVisitEndpoint endpoint.Endpoint
//  }
//
//  func (e *Endpoints) CreateVisit(ctx context.Context, visit *entity.Visit) (res *entity.Visit, err error) {
//  	req := CreateVisitRequest{
//  		Visit: visit,
//  	}
//  	resp, err := e.CreateVisitEndpoint(ctx, &req)
//  	if err != nil {
//  		return
//  	}
//  	return resp.(*CreateVisitResponse).Res, resp.(*CreateVisitResponse).Err
//  }
//
//  func CreateVisitEndpoint(svc VisitService) endpoint.Endpoint {
//  	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
//  		req := request.(*CreateVisitRequest)
//  		res, err := svc.CreateVisit(ctx, req.Visit)
//  		return &CreateVisitResponse{
//  			Res: res,
//  			Err: err,
//  		}, nil
//  	}
//  }
//
func (EndpointsTemplate) Render(i *parser.Interface) *File {
	f := NewFile(i.PackageName)

	f.Type().Id("Endpoints").StructFunc(func(g *Group) {
		for _, signature := range i.FuncSignatures {
			g.Id(signature.Name+"Endpoint").Qual(PackageAliasGoKit, "Endpoint")
		}
	})

	for _, signature := range i.FuncSignatures {
		f.Add(endpointFunc(signature))
	}
	f.Line()
	for _, signature := range i.FuncSignatures {
		f.Add(newEndpointFunc(signature, i))
	}

	return f
}

func (EndpointsTemplate) Path() string {
	return "./endpoints.go"
}

func endpointFunc(signature *parser.FuncSignature) *Statement {
	return methodDefinition("Endpoints", signature).
		BlockFunc(endpointBody(signature))
}

func endpointBody(signature *parser.FuncSignature) func(g *Group) {
	req := "req"
	resp := "resp"
	return func(g *Group) {
		g.Id(req).Op(":=").Id(signature.Name + "Request").Values(mapInitByFuncFields(signature.Params))
		g.List(Id(resp), Err()).Op(":=").Id(util.FirstLowerChar("Endpoint")).Dot(signature.Name+"Endpoint").Call(Id(Context), Op("&").Id(req))
		g.If(Err().Op("!=").Nil()).Block(
			Return(),
		)
		g.ReturnFunc(func(group *Group) {
			for _, field := range signature.Results {
				group.Add(typeCasting("", resp, signature.Name+"Response")).Op(".").Add(structFieldName(field))
			}
		})
	}
}

func newEndpointFuncBody(signature *parser.FuncSignature) *Statement {
	return Return(Func().Params(
		Id("ctx").Qual("context", "Context"),
		Id("request").Interface(),
	).Params(
		Interface(),
		Error(),
	).BlockFunc(func(g *Group) {
		g.Add(typeCasting("req", "request", signature.Name+"Request"))
		g.Add(fullServiceMethodCall("svc", "req", signature))
		g.Return(
			Op("&").Id(signature.Name+"Response").Values(mapInitByFuncFields(signature.Results)),
			Nil(),
		)
	}))
}

func newEndpointFunc(signature *parser.FuncSignature, svcInterface *parser.Interface) *Statement {
	return Func().
		Id(signature.Name + "Endpoint").Params(Id("svc").Id(svcInterface.Name)).Params(Qual(PackageAliasGoKit, "Endpoint")).
		Block(newEndpointFuncBody(signature))
}
