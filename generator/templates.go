package generator

import "github.com/devimteam/microgen/parser"

type RenderData struct {
	PackageFullName string
	Interface       *parser.Interface
}

var ExchangeTemplate = Template{
	TemplatePath: "generator/templates/exchange.go.tmpl",
	ResultPath:   "exchange.go",
}

var EndpointTemplate = Template{
	TemplatePath: "generator/templates/endpoints.go.tmpl",
	ResultPath:   "endpoints.go",
}

var ClientTemplate = Template{
	TemplatePath: "generator/templates/client.go.tmpl",
	ResultPath:   "client.go",
}

var LoggingMiddlewareTemplate = Template{
	TemplatePath: "generator/templates/middleware/logging.go.tmpl",
	ResultPath:   "middleware/endpoints.go",
}
