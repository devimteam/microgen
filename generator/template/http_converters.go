package template

import (
	"path/filepath"

	"github.com/devimteam/microgen/generator/write_strategy"
	"github.com/devimteam/microgen/util"
	"github.com/vetcher/godecl/types"
	. "github.com/vetcher/jennifer/jen"
)

const (
	defaultResponseEncoderName = "DefaultResponseEncoder"
	defaultRequestEncoderName  = "DefaultRequestEncoder"
)

type httpConverterTemplate struct {
	Info                          *GenerationInfo
	encodersRequest               []*types.Function
	decodersRequest               []*types.Function
	encodersResponse              []*types.Function
	decodersResponse              []*types.Function
	state                         WriteStrategyState
	isDefaultEncoderRequestExist  bool
	isDefaultEncoderResponseExist bool
}

func NewHttpConverterTemplate(info *GenerationInfo) Template {
	return &httpConverterTemplate{
		Info: info.Duplicate(),
	}
}

func (t *httpConverterTemplate) DefaultPath() string {
	return "./transport/converter/http/exchange_converters.go"
}

func (t *httpConverterTemplate) ChooseStrategy() (write_strategy.Strategy, error) {
	if err := util.TryToOpenFile(t.Info.AbsOutPath, t.DefaultPath()); t.Info.Force || err != nil {
		t.state = FileStrat
		return write_strategy.NewCreateFileStrategy(t.Info.AbsOutPath, t.DefaultPath()), nil
	}
	file, err := util.ParseFile(filepath.Join(t.Info.AbsOutPath, t.DefaultPath()))
	if err != nil {
		return nil, err
	}

	removeAlreadyExistingFunctions(file.Functions, &t.encodersRequest, httpEncodeRequestName)
	removeAlreadyExistingFunctions(file.Functions, &t.decodersRequest, httpDecodeRequestName)
	removeAlreadyExistingFunctions(file.Functions, &t.encodersResponse, httpEncodeResponseName)
	removeAlreadyExistingFunctions(file.Functions, &t.decodersResponse, httpDecodeResponseName)

	for i := range file.Functions {
		if file.Functions[i].Name == defaultResponseEncoderName {
			t.isDefaultEncoderResponseExist = true
			continue
		}
		if file.Functions[i].Name == defaultRequestEncoderName {
			t.isDefaultEncoderRequestExist = true
			break
		}
		if t.isDefaultEncoderRequestExist && t.isDefaultEncoderResponseExist {
			break
		}
	}

	t.state = AppendStrat
	return write_strategy.NewAppendToFileStrategy(t.Info.AbsOutPath, t.DefaultPath()), nil
}

func (t *httpConverterTemplate) Prepare() error {
	for _, fn := range t.Info.Iface.Methods {
		t.decodersRequest = append(t.decodersRequest, fn)
		t.encodersRequest = append(t.encodersRequest, fn)
		t.decodersResponse = append(t.decodersResponse, fn)
		t.encodersResponse = append(t.encodersResponse, fn)
	}
	return nil
}

func (t *httpConverterTemplate) Render() write_strategy.Renderer {
	f := &Statement{}

	if !t.isDefaultEncoderRequestExist {
		f.Line().Add(defaultEncoderRequest()).Line()
	}
	if !t.isDefaultEncoderResponseExist {
		f.Line().Add(defaultEncoderResponse()).Line()
	}

	for _, fn := range t.decodersRequest {
		f.Line().Add(t.decodeHttpRequest(fn)).Line()
	}
	for _, fn := range t.decodersResponse {
		f.Line().Add(t.decodeHttpResponse(fn)).Line()
	}
	for _, fn := range t.encodersRequest {
		f.Line().Add(encodeHttpRequest(fn)).Line()
	}
	for _, fn := range t.encodersResponse {
		f.Line().Add(encodeHttpResponse(fn)).Line()
	}

	if t.state == AppendStrat {
		return f
	}

	file := NewFile("httpconv")
	file.PackageComment(FileHeader)
	file.PackageComment(`Please, do not change functions names!`)
	file.Add(f)

	return file
}

// https://github.com/go-kit/kit/blob/master/examples/addsvc/pkg/addtransport/http.go#L201
func defaultEncoderRequest() *Statement {
	return Func().Id(defaultRequestEncoderName).
		Params(
			Id("_").Qual(PackagePathContext, "Context"),
			Id("r").Op("*").Qual(PackagePathHttp, "Request"),
			Id("request").Interface(),
		).Params(
		Error(),
	).BlockFunc(func(g *Group) {
		g.Var().Id("buf").Qual(PackagePathBytes, "Buffer")
		g.If(
			Err().Op(":=").Qual(PackagePathJson, "NewEncoder").Call(Op("&").Id("buf")).Dot("Encode").Call(Id("request")),
			Err().Op("!=").Nil(),
		).Block(
			Return(Err()),
		)
		g.Id("r").Dot("Body").Op("=").Qual(PackagePathIOUtil, "NopCloser").Call(Op("&").Id("buf"))
		g.Return(Nil())
	})
}

// https://github.com/go-kit/kit/blob/master/examples/addsvc/pkg/addtransport/http.go#L212
func defaultEncoderResponse() *Statement {
	return Func().Id(defaultResponseEncoderName).
		Params(
			Id("_").Qual(PackagePathContext, "Context"),
			Id("w").Qual(PackagePathHttp, "ResponseWriter"),
			Id("response").Interface(),
		).Params(
		Error(),
	).BlockFunc(func(g *Group) {
		g.Id("w").Dot("Header").Call().Dot("Set").Call(Lit("Content-Type"), Lit("application/json; charset=utf-8"))
		g.Return(
			Qual(PackagePathJson, "NewEncoder").Call(Id("w")).Dot("Encode").Call(Id("response")),
		)
	})
}

func (t *httpConverterTemplate) decodeHttpRequest(fn *types.Function) *Statement {
	return Func().Id(httpDecodeRequestName(fn)).
		Params(
			Id("_").Qual(PackagePathContext, "Context"),
			Id("r").Op("*").Qual(PackagePathHttp, "Request"),
		).Params(
		Interface(),
		Error(),
	).BlockFunc(func(g *Group) {
		g.Var().Id("req").Qual(t.Info.ServiceImportPath, requestStructName(fn))
		g.Err().Op(":=").Qual(PackagePathJson, "NewDecoder").Call(Id("r").Dot("Body")).Dot("Decode").Call(Op("&").Id("req"))
		g.Return(Id("req"), Err())
	})
}

func (t *httpConverterTemplate) decodeHttpResponse(fn *types.Function) *Statement {
	return Func().Id(httpDecodeResponseName(fn)).
		Params(
			Id("_").Qual(PackagePathContext, "Context"),
			Id("r").Op("*").Qual(PackagePathHttp, "Response"),
		).Params(
		Interface(),
		Error(),
	).
		BlockFunc(func(g *Group) {
			g.Var().Id("resp").Qual(t.Info.ServiceImportPath, responseStructName(fn))
			g.Err().Op(":=").Qual(PackagePathJson, "NewDecoder").Call(Id("r").Dot("Body")).Dot("Decode").Call(Op("&").Id("resp"))
			g.Return(Id("resp"), Err())
		})
}

func encodeHttpResponse(fn *types.Function) *Statement {
	return Func().Id(httpEncodeResponseName(fn)).Params(
		Id("ctx").Qual(PackagePathContext, "Context"),
		Id("w").Qual(PackagePathHttp, "ResponseWriter"),
		Id("response").Interface(),
	).Params(
		Error(),
	).Block(
		Return().Id(defaultResponseEncoderName).Call(Id("ctx"), Id("w"), Id("response")),
	)
}

func encodeHttpRequest(fn *types.Function) *Statement {
	return Func().Id(httpEncodeRequestName(fn)).Params(
		Id("ctx").Qual(PackagePathContext, "Context"),
		Id("r").Op("*").Qual(PackagePathHttp, "Request"),
		Id("request").Interface(),
	).Params(
		Error(),
	).Block(
		Return().Id(defaultRequestEncoderName).Call(Id("ctx"), Id("r"), Id("request")),
	)
}

func httpDecodeRequestName(f *types.Function) string {
	return "DecodeHTTP" + util.ToUpperFirst(f.Name) + "Request"
}

func httpEncodeRequestName(f *types.Function) string {
	return "EncodeHTTP" + util.ToUpperFirst(f.Name) + "Request"
}

func httpEncodeResponseName(f *types.Function) string {
	return "EncodeHTTP" + util.ToUpperFirst(f.Name) + "Response"
}

func httpDecodeResponseName(f *types.Function) string {
	return "DecodeHTTP" + util.ToUpperFirst(f.Name) + "Response"
}
