package template

import (
	. "github.com/dave/jennifer/jen"
	"github.com/devimteam/microgen/parser"
	"github.com/devimteam/microgen/util"
)

const (
	RequestConvType  = "request"
	ResponseConvType = "response"
)

var (
	defaultProtobufTypes = []string{"string", "[]byte", "bool", "int32", "int64", "uint32", "uint64", "float32", "float64"}
	goToProtobufTypesMap = map[string]string{
		"uint": "uint64",
		"int":  "int64",
	}
)

type GRPCConverterTemplate struct {
	PackagePath string
}

func utilPackagePath(path string) string {
	return path + "/util"
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
func (t GRPCConverterTemplate) Render(i *parser.Interface) *File {
	f := NewFile("transportgrpc")

	for _, signature := range i.FuncSignatures {
		f.Var().Id(converterStructName(signature)).Op("=").Op("&").Qual(PackagePathTransportLayerGRPC, "EndpointConverter").
			ValuesFunc(func(g *Group) {
				g.Add(t.converterFunc(signature, signature.Params, i, RequestConvType, false))
				g.Add(t.converterFunc(signature, signature.Results, i, ResponseConvType, false))
				g.Add(t.converterFunc(signature, signature.Params, i, RequestConvType, true))
				g.Add(t.converterFunc(signature, signature.Results, i, ResponseConvType, true))
				g.Line()
			})
	}

	return f
}

func nameToProto(name string, reverse bool) string {
	if !reverse {
		return name + "ToProto"
	}
	return "ToProto" + name
}

func (GRPCConverterTemplate) Path() string {
	return "./transport/grpc/converter.go"
}

// Renders exchanges that represents requests and responses.
//
//  type CreateVisitRequest struct {
//  	Visit *entity.Visit `json:"visit"`
//  }
//
func detectCustomType(object string, field *parser.FuncField) (Code, bool) {
	if isDefaultProtobufType(field.Type) {
		return Id(object).Dot(util.ToUpperFirst(field.Name)), false
	}
	if newType, ok := goToProtobufTypesMap[field.Type]; ok {
		return Id(object).Dot(util.ToUpperFirst(field.Name)).Assert(Id(newType)), false
	}
	return Add(), true
}

func isDefaultProtobufType(typeName string) bool {
	for _, t := range defaultProtobufTypes {
		if t == typeName {
			return true
		}
	}
	return false
}

func (t GRPCConverterTemplate) converterFunc(signature *parser.FuncSignature, fields []*parser.FuncField, i *parser.Interface, convType string, reverse bool) Code {
	var structName string
	if convType == RequestConvType {
		structName = requestStructName(signature)
	} else if convType == ResponseConvType {
		structName = responseStructName(signature)
	}
	return Line().Func().Call(Op("_").Qual(PackagePathContext, "Context"), Id("data").Interface()).Params(Interface(), Error()).BlockFunc(
		func(group *Group) {
			group.Id(convType).Op(":=").Id("data").Assert(Op("*").Qual(t.PackagePath, structName))
			group.Return().List(Op("&").Qual(protobufPath(i), structName).Values(DictFunc(func(dict Dict) {
				for _, field := range removeContextIfFirst(fields) {
					code, ok := detectCustomType(convType, field)
					if ok {
						code = Qual(utilPackagePath(t.PackagePath), nameToProto(util.ToUpperFirst(field.Name), reverse)).
							Call(Id(convType).
								Dot(util.ToUpperFirst(field.Name)))
					}
					dict[structFieldName(field)] = Line().Add(code)
				}
			})), Nil())
		},
	)
}
