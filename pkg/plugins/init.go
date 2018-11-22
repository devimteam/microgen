package plugins

import "github.com/devimteam/microgen/pkg/microgen"

func init() {
	microgen.RegisterPlugin(LoggingPlugin, &loggingMiddlewarePlugin{})
	microgen.RegisterPlugin(RecoveringPlugin, &recoveringMiddlewarePlugin{})
	microgen.RegisterPlugin(opentracingPlugin, &opentracingMiddlewarePlugin{})
	microgen.RegisterPlugin(transportKitPlugin, &transportGokitPlugin{})
	microgen.RegisterPlugin(grpcKitPlugin, &grpcGokitPlugin{})
}

const (
	serviceAlias = "service"
	_service_    = "svc"
	_logger_     = "logger"
	_ctx_        = "ctx"
	_next_       = "next"
	_Request_    = "Request"
	_Response_   = "Response"
	_i_          = "i"
	_tracer_     = "tracer"
)
