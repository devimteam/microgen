package template

import (
	"context"
	"path/filepath"

	. "github.com/dave/jennifer/jen"
	"github.com/devimteam/microgen/generator/write_strategy"
	"github.com/vetcher/go-astra/types"
)

type jsonrpcEndpointConverterTemplate struct {
	info             *GenerationInfo
	requestEncoders  []*types.Function
	requestDecoders  []*types.Function
	responseEncoders []*types.Function
	responseDecoders []*types.Function
	state            WriteStrategyState
}

func NewJSONRPCEndpointConverterTemplate(info *GenerationInfo) Template {
	return &jsonrpcEndpointConverterTemplate{
		info: info,
	}
}

func (t *jsonrpcEndpointConverterTemplate) Render(ctx context.Context) write_strategy.Renderer {
	f := &Statement{}

	for _, signature := range t.requestEncoders {
		f.Line().Add(t.encodeRequest(signature))
	}
	for _, signature := range t.responseEncoders {
		f.Line().Add(t.encodeResponse(signature))
	}
	for _, signature := range t.requestDecoders {
		f.Line().Add(t.decodeRequest(signature))
	}
	for _, signature := range t.responseDecoders {
		f.Line().Add(t.decodeResponse(signature))
	}

	if t.state == AppendStrat {
		return f
	}

	file := NewFile("jsonrpcconv")
	file.ImportAlias(t.info.SourcePackageImport, serviceAlias)
	file.HeaderComment(t.info.FileHeader)
	file.PackageComment(`Please, do not change functions names!`)
	file.Add(f)

	return file
}

func (jsonrpcEndpointConverterTemplate) DefaultPath() string {
	return "./transport/converter/jsonrpc/exchange_converters.go"
}

func (t *jsonrpcEndpointConverterTemplate) Prepare(ctx context.Context) error {
	for _, fn := range t.info.Iface.Methods {
		t.requestDecoders = append(t.requestDecoders, fn)
		t.requestEncoders = append(t.requestEncoders, fn)
		t.responseDecoders = append(t.responseDecoders, fn)
		t.responseEncoders = append(t.responseEncoders, fn)
	}
	return nil
}

func (t *jsonrpcEndpointConverterTemplate) ChooseStrategy(ctx context.Context) (write_strategy.Strategy, error) {
	if err := statFile(t.info.OutputFilePath, t.DefaultPath()); err != nil {
		t.state = FileStrat
		return write_strategy.NewCreateFileStrategy(t.info.OutputFilePath, t.DefaultPath()), nil
	}
	file, err := parsePackage(filepath.Join(t.info.OutputFilePath, t.DefaultPath()))
	if err != nil {
		return nil, err
	}

	removeAlreadyExistingFunctions(file.Functions, &t.requestEncoders, encodeRequestName)
	removeAlreadyExistingFunctions(file.Functions, &t.requestDecoders, decodeRequestName)
	removeAlreadyExistingFunctions(file.Functions, &t.responseEncoders, encodeResponseName)
	removeAlreadyExistingFunctions(file.Functions, &t.responseDecoders, decodeResponseName)

	t.state = AppendStrat
	return write_strategy.NewAppendToFileStrategy(t.info.OutputFilePath, t.DefaultPath()), nil
}

func (t *jsonrpcEndpointConverterTemplate) encodeRequest(fn *types.Function) Code {
	fullName := "request"
	return Line().Func().Id(encodeRequestName(fn)).Params(Op("_").Qual(PackagePathContext, "Context"), Id(fullName).Interface()).
		Params(Qual(PackagePathJson, "RawMessage"), Error()).BlockFunc(
		func(group *Group) {
			group.Return().Qual(PackagePathJson, "Marshal").Call(Id(fullName))
		})
}

func (t *jsonrpcEndpointConverterTemplate) encodeResponse(fn *types.Function) Code {
	fullName := "response"
	return Line().Func().Id(encodeResponseName(fn)).Params(Op("_").Qual(PackagePathContext, "Context"), Id(fullName).Interface()).
		Params(Qual(PackagePathJson, "RawMessage"), Error()).BlockFunc(
		func(group *Group) {
			group.Return().Qual(PackagePathJson, "Marshal").Call(Id(fullName))
		})
}

func (t *jsonrpcEndpointConverterTemplate) decodeRequest(fn *types.Function) Code {
	fullName := "request"
	shortName := "req"
	return Line().Func().Id(decodeRequestName(fn)).Params(Op("_").Qual(PackagePathContext, "Context"), Id(fullName).Qual(PackagePathGoKitTransportJSONRPC, "Response")).
		Params(Interface(), Error()).BlockFunc(
		func(group *Group) {
			group.If(Id(fullName).Dot("Error").Op("!=").Nil()).Block(
				Return(Nil(), Id(fullName).Dot("Error")),
			)
			group.Var().Id(shortName).Qual(t.info.SourcePackageImport, requestStructName(fn))
			group.Err().Op(":=").Qual(PackagePathJson, "Unmarshal").Call(Id(fullName), Op("&").Id(shortName))
			group.Return(Op("&").Id(shortName), Err())
		})
}

func (t *jsonrpcEndpointConverterTemplate) decodeResponse(fn *types.Function) Code {
	fullName := "response"
	shortName := "resp"
	return Line().Func().Id(decodeResponseName(fn)).Params(Op("_").Qual(PackagePathContext, "Context"), Id(fullName).Qual(PackagePathGoKitTransportJSONRPC, "Response")).
		Params(Interface(), Error()).BlockFunc(
		func(group *Group) {
			group.If(Id(fullName).Dot("Error").Op("!=").Nil()).Block(
				Return(Nil(), Id(fullName).Dot("Error")),
			)
			group.Var().Id(shortName).Qual(t.info.SourcePackageImport, responseStructName(fn))
			group.Err().Op(":=").Qual(PackagePathJson, "Unmarshal").Call(Id(fullName), Op("&").Id(shortName))
			group.Return(Op("&").Id(shortName), Err())
		})
}
