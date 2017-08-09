package template

import (
	. "github.com/dave/jennifer/jen"
	"github.com/devimteam/microgen/parser"
)

type ExchangeTemplate struct {
}

func requestStructName(signature *parser.FuncSignature) string {
	return signature.Name + "Request"
}

func responseStructName(signature *parser.FuncSignature) string {
	return signature.Name + "Response"
}

// Renders exchanges file.
//
//  package visitsvc
//
//  import (
//  	"gitlab.devim.team/microservices/visitsvc/entity"
//  )
//
//  type CreateVisitRequest struct {
//  	Visit *entity.Visit `json:"visit"`
//  }
//
//  type CreateVisitResponse struct {
//  	Res *entity.Visit `json:"res"`
//  	Err error         `json:"err"`
//  }
//
func (ExchangeTemplate) Render(i *parser.Interface) *File {
	f := NewFile(i.PackageName)

	for _, signature := range i.FuncSignatures {
		f.Add(exchange(requestStructName(signature), signature.Params))
		f.Add(exchange(responseStructName(signature), signature.Results))
	}

	return f
}

func (ExchangeTemplate) Path() string {
	return "./exchanges.go"
}

// Renders exchanges that represents requests and responses.
//
//  type CreateVisitRequest struct {
//  	Visit *entity.Visit `json:"visit"`
//  }
//
func exchange(name string, params []*parser.FuncField) Code {
	return Type().Id(name).StructFunc(func(g *Group) {
		for i, param := range params {

			// skip "context" package entry if it is first arg
			if i == 0 && isContext(param) {
				continue
			}

			g.Add(structField(param))
		}
	}).Line()
}
