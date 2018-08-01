package stringsvc

import (
	"context"
)

// @microgen middleware, logging, grpc, http, recovering, error-logging, tracing, caching, metrics, service-discovery
// @grpc-addr service.string.StringService
// @protobuf github.com/devimteam/microgen/examples/protobuf
type StringService interface {
	// @logs-ignore ans, err
	// @cache
	Uppercase(ctx context.Context, stringsMap map[string]string) (ans string, err error)
	// @http-method geT
	// @cache-key text
	// @json-rpc-prefix v1.
	Count(ctx context.Context, text string, symbol string) (count int, positions []int, err error)
	// @logs-len comments
	TestCase(ctx context.Context, comments []*Comment) (tree map[string]int, err error)

	DummyMethod(ctx context.Context) (err error)

	// @microgen -
	IgnoredMethod()
	// @microgen -
	IgnoredErrorMethod() error
}
