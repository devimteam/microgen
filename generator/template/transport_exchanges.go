package template

import (
	"context"

	. "github.com/dave/jennifer/jen"
	"github.com/devimteam/microgen/generator/write_strategy"
	"github.com/vetcher/go-astra/types"
)

type exchangeTemplate struct {
	info *GenerationInfo
}

func NewExchangeTemplate(info *GenerationInfo) Template {
	return &exchangeTemplate{
		info: info,
	}
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
func (t *exchangeTemplate) Render(ctx context.Context) write_strategy.Renderer {
	f := NewFile("transport")
	f.HeaderComment(t.info.FileHeader)

	if len(t.info.Iface.Methods) > 0 {
		f.Type().Op("(")
	}
	for _, signature := range t.info.Iface.Methods {
		if t.info.AllowedMethods[signature.Name] {
			f.Add(exchange(ctx, requestStructName(signature), RemoveContextIfFirst(signature.Args))) //.Line()
			f.Add(exchange(ctx, responseStructName(signature), removeErrorIfLast(signature.Results))).Line()
		}
	}
	if len(t.info.Iface.Methods) > 0 {
		f.Op(")")
	}

	return f
}

func (exchangeTemplate) DefaultPath() string {
	return filenameBuilder(PathTransport, "exchanges")
}

func (exchangeTemplate) Prepare(ctx context.Context) error {
	return nil
}

func (t *exchangeTemplate) ChooseStrategy(ctx context.Context) (write_strategy.Strategy, error) {
	return write_strategy.NewCreateFileStrategy(t.info.OutputFilePath, t.DefaultPath()), nil
}

// Renders exchanges that represents requests and responses.
//
//  type CreateVisitRequest struct {
//  	Visit *entity.Visit `json:"visit"`
//  }
//
func exchange(ctx context.Context, name string, params []types.Variable) Code {
	if len(params) == 0 {
		return Comment("Formal exchange type, please do not delete.").Line().
			Id(name).Struct()
		//Line()
	}
	return Id(name).StructFunc(func(g *Group) {
		for _, param := range params {
			g.Add(structField(ctx, &param))
		}
	}) //.Line()
}
