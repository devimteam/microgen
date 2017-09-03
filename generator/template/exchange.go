package template

import (
	"github.com/vetcher/godecl/types"
	. "github.com/vetcher/jennifer/jen"
)

type ExchangeTemplate struct {
	PkgName string
}

func requestStructName(signature *types.Function) string {
	return signature.Name + "Request"
}

func responseStructName(signature *types.Function) string {
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
func (t *ExchangeTemplate) Render(i *types.Interface) *Statement {
	f := Statement{}

	for _, signature := range i.Methods {
		f.Add(exchange(requestStructName(signature), signature.Args)).Line()
		f.Add(exchange(responseStructName(signature), signature.Results)).Line()
	}

	return &f
}

func (ExchangeTemplate) Path() string {
	return "./exchanges.go"
}

func (t *ExchangeTemplate) PackageName() string {
	return t.PkgName
}

// Renders exchanges that represents requests and responses.
//
//  type CreateVisitRequest struct {
//  	Visit *entity.Visit `json:"visit"`
//  }
//
func exchange(name string, params []types.Variable) Code {
	return Type().Id(name).StructFunc(func(g *Group) {
		for _, param := range removeContextIfFirst(params) {
			g.Add(structField(&param))
		}
	}).Line()
}
