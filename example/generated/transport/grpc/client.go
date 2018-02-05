// This file was automatically generated by "microgen 0.8.0alpha" utility.
// Please, do not edit.
package transportgrpc

import (
	generated "github.com/devimteam/microgen/example/generated"
	protobuf "github.com/devimteam/microgen/example/generated/transport/converter/protobuf"
	protobuf1 "github.com/devimteam/microgen/example/protobuf"
	log "github.com/go-kit/kit/log"
	opentracing "github.com/go-kit/kit/tracing/opentracing"
	grpc1 "github.com/go-kit/kit/transport/grpc"
	opentracinggo "github.com/opentracing/opentracing-go"
	grpc "google.golang.org/grpc"
)

func NewGRPCClient(conn *grpc.ClientConn, logger log.Logger, tracer opentracinggo.Tracer, opts ...grpc1.ClientOption) generated.StringService {
	opts = append(opts, grpc1.ClientBefore(
		opentracing.ContextToGRPC(tracer, logger),
	))
	return &generated.Endpoints{
		CountEndpoint: opentracing.TraceClient(
			tracer,
			"Count",
		)(
			grpc1.NewClient(
				conn,
				"service.string",
				"Count",
				protobuf.EncodeCountRequest,
				protobuf.DecodeCountResponse,
				protobuf1.CountResponse{},
				opts...,
			).Endpoint(),
		),
		TestCaseEndpoint: opentracing.TraceClient(
			tracer,
			"TestCase",
		)(
			grpc1.NewClient(
				conn,
				"service.string",
				"TestCase",
				protobuf.EncodeTestCaseRequest,
				protobuf.DecodeTestCaseResponse,
				protobuf1.TestCaseResponse{},
				opts...,
			).Endpoint(),
		),
		UppercaseEndpoint: opentracing.TraceClient(
			tracer,
			"Uppercase",
		)(
			grpc1.NewClient(
				conn,
				"service.string",
				"Uppercase",
				protobuf.EncodeUppercaseRequest,
				protobuf.DecodeUppercaseResponse,
				protobuf1.UppercaseResponse{},
				opts...,
			).Endpoint(),
		),
	}
}
