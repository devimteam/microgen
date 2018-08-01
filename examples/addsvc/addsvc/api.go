package addsvc

import "context"

// Service describes a service that adds things together.
// Service realization should be in addsvc/service package.
//
// @microgen middleware, logging, grpc, http, recovering, error-logging, tracing, caching, metrics
// @protobuf github.com/devimteam/microgen/examples/protobuf
type Service interface {
	Sum(ctx context.Context, a, b int) (result int, err error)
	Concat(ctx context.Context, a, b string) (result string, err error)
}
