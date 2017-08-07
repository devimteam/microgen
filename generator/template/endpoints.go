package template

import (
	. "github.com/dave/jennifer/jen"
	"github.com/devimteam/microgen/parser"
)

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
			g.Id(signature.Name+"Endpoint").Qual("github.com/go-kit/kit/endpoint", "Endpoint")
		}
	})

	for _, signature := range i.FuncSignatures {
		endpointFunc(signature)
	}

	return f
}

func (EndpointsTemplate) Path() string {
	return "./endpoints.go"
}

func endpointFunc(signature *parser.FuncSignature) Code {
	return Func().
		Params(Id("e").Op("*").Id("Endpoints")).
		Id(signature.Name).
		Params(funcParams(signature.Params)).
		Params(funcParams(signature.Results)).
		BlockFunc(
			func(g *Group) {
				Return()
				//g.Id("req").Op(":=").Id(signature.Name+"Request").StructFunc(func(g *Group) {
				//
				//})
			})
}
