package plugins

import "github.com/devimteam/microgen/pkg/microgen"

func init() {
	microgen.RegisterPlugin(loggingPlugin, &loggingMiddlewarePlugin{})
	microgen.RegisterPlugin(recoveringPlugin, &recoveringMiddlewarePlugin{})
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
)
