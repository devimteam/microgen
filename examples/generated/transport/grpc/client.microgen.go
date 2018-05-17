// Code generated by microgen 1.0.0-alpha. DO NOT EDIT.

package transportgrpc

import (
	transport "github.com/devimteam/microgen/examples/generated/transport"
	pb "github.com/devimteam/microgen/examples/protobuf"
	log "github.com/go-kit/kit/log"
	opentracing "github.com/go-kit/kit/tracing/opentracing"
	grpckit "github.com/go-kit/kit/transport/grpc"
	opentracinggo "github.com/opentracing/opentracing-go"
	grpc "google.golang.org/grpc"
)

func NewGRPCClient(conn *grpc.ClientConn, addr string, opts ...grpckit.ClientOption) transport.EndpointsSet {
	return transport.EndpointsSet{
		CountEndpoint: grpckit.NewClient(
			conn, addr, "Count",
			_Encode_Count_Request,
			_Decode_Count_Response,
			pb.CountResponse{},
			opts...,
		).Endpoint(),
		TestCaseEndpoint: grpckit.NewClient(
			conn, addr, "TestCase",
			_Encode_TestCase_Request,
			_Decode_TestCase_Response,
			pb.TestCaseResponse{},
			opts...,
		).Endpoint(),
		UppercaseEndpoint: grpckit.NewClient(
			conn, addr, "Uppercase",
			_Encode_Uppercase_Request,
			_Decode_Uppercase_Response,
			pb.UppercaseResponse{},
			opts...,
		).Endpoint(),
	}
}

func TracingGRPCClientOptions(tracer opentracinggo.Tracer, logger log.Logger) func([]grpckit.ClientOption) []grpckit.ClientOption {
	return func(opts []grpckit.ClientOption) []grpckit.ClientOption {
		return append(opts, grpckit.ClientBefore(
			opentracing.ContextToGRPC(tracer, logger),
		))
	}
}