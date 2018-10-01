package plugins

import "github.com/devimteam/microgen/pkg/microgen"

func init() {
	microgen.RegisterPlugin(loggingPlugin, &loggingMiddlewarePlugin{})
}
