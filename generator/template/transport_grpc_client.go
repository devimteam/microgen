package template

import (
	"context"
	"errors"

	. "github.com/dave/jennifer/jen"
	"github.com/devimteam/microgen/generator/write_strategy"
	"github.com/vetcher/go-astra/types"
)

var (
	ErrProtobufEmpty = errors.New("protobuf package is empty")
)

type gRPCClientTemplate struct {
	Info *GenerationInfo
}

func NewGRPCClientTemplate(info *GenerationInfo) Template {
	return &gRPCClientTemplate{
		Info: info,
	}
}

func (t *gRPCClientTemplate) grpcConverterPackagePath() string {
	return t.Info.SourcePackageImport + "/transport/converter/protobuf"
}

// Render whole grpc client file.
//
//		// This file was automatically generated by "microgen" utility.
//		// DO NOT EDIT.
//		package transportgrpc
//
//		import (
//			svc "github.com/devimteam/microgen/example/svc"
//			protobuf "github.com/devimteam/microgen/example/svc/transport/converter/protobuf"
//			grpc1 "github.com/go-kit/kit/transport/grpc"
//			stringsvc "gitlab.devim.team/protobuf/stringsvc"
//			grpc "google.golang.org/grpc"
//		)
//
//		func NewGRPCClient(conn *grpc.ClientConn, opts ...grpc1.ClientOption) svc.StringService {
//			return &svc.Endpoints{CountEndpoint: grpc1.NewClient(
//				conn,
//				"devim.string.protobuf.StringService",
//				"Count",
//				protobuf.EncodeCountRequest,
//				protobuf.DecodeCountResponse,
//				stringsvc.CountResponse{},
//				opts...,
//			).Endpoint()}
//		}
//
func (t *gRPCClientTemplate) Render(ctx context.Context) write_strategy.Renderer {
	f := NewFile("transportgrpc")
	f.ImportAlias(t.Info.ProtobufPackageImport, "pb")
	f.ImportAlias(t.Info.SourcePackageImport, serviceAlias)
	f.PackageComment(t.Info.FileHeader)

	f.Func().Id("NewGRPCClient").
		ParamsFunc(func(p *Group) {
			p.Id("conn").Op("*").Qual(PackagePathGoogleGRPC, "ClientConn")
			p.Id("addr").Id("string")
			if Tags(ctx).Has(TracingTag) {
				p.Id("logger").Qual(PackagePathGoKitLog, "Logger")
			}
			p.Id("opts").Op("...").Qual(PackagePathGoKitTransportGRPC, "ClientOption")
		}).Qual(t.Info.SourcePackageImport, EndpointsSetName).
		BlockFunc(func(g *Group) {
			if Tags(ctx).Has(TracingTag) {
				g.Id("opts").Op("=").Append(Id("opts"), Qual(PackagePathGoKitTransportGRPC, "ClientBefore").Call(
					Line().Qual(PackagePathGoKitTracing, "ContextToGRPC").Call(Id("tracer"), Id("logger")).Op(",").Line(),
				))
			}
			g.Return().Qual(t.Info.SourcePackageImport, EndpointsSetName).Values(DictFunc(func(d Dict) {
				for _, m := range t.Info.Iface.Methods {
					client := &Statement{}
					client.Qual(PackagePathGoKitTransportGRPC, "NewClient").Call(
						Line().Id("conn"),
						Line().Id("addr"),
						Line().Lit(m.Name),
						Line().Qual(pathToConverter(t.Info.SourcePackageImport), encodeRequestName(m)),
						Line().Qual(pathToConverter(t.Info.SourcePackageImport), decodeResponseName(m)),
						Line().Add(t.replyType(m)),
						Line().Add(t.clientOpts(m)).Op("...").Line(),
					).Dot("Endpoint").Call()
					d[Id(endpointStructName(m.Name))] = client
				}
			}))
		})
	return f
}

// Renders reply type argument
// 		stringsvc.CountResponse{}
func (t *gRPCClientTemplate) replyType(signature *types.Function) *Statement {
	results := removeErrorIfLast(signature.Results)
	if len(results) == 0 {
		return Qual(PackagePathEmptyProtobuf, "Empty").Values()
	}
	if len(results) == 1 {
		sp := specialReplyType(results[0].Type)
		if sp != nil {
			return sp
		}
	}
	return Qual(t.Info.ProtobufPackageImport, responseStructName(signature)).Values()
}

func specialReplyType(p types.Type) *Statement {
	name := types.TypeName(p)
	imp := types.TypeImport(p)
	// *string -> *wrappers.StringValue
	if name != nil && *name == "string" && imp == nil {
		ptr, ok := p.(types.TPointer)
		if ok && ptr.NumberOfPointers == 1 {
			return (&Statement{}).Qual(GolangProtobufWrappers, "StringValue").Values()
		}
	}
	return nil
}

func (gRPCClientTemplate) DefaultPath() string {
	return filenameBuilder(PathTransport, "grpc", "client")
}

func (t *gRPCClientTemplate) Prepare(ctx context.Context) error {
	if t.Info.ProtobufPackageImport == "" {
		return ErrProtobufEmpty
	}
	return nil
}

func (t *gRPCClientTemplate) ChooseStrategy(ctx context.Context) (write_strategy.Strategy, error) {
	return write_strategy.NewCreateFileStrategy(t.Info.AbsOutputFilePath, t.DefaultPath()), nil
}

func (t *gRPCClientTemplate) clientOpts(fn *types.Function) *Statement {
	s := &Statement{}
	s.Id("opts")
	return s
}
