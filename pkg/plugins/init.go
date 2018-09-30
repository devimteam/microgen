package plugins

import "microgen/pkg/microgen"

func init() {
	microgen.RegisterPlugin(loggingPlugin, &loggingMiddlewarePlugin{})
}
