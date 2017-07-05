package generator

var RequestsTemplate = Template{
	TemplatePath: "generator/templates/requests.go.tmpl",
	ResultPath:   "requests.go",
}

var ResponsesTemplate = Template{
	TemplatePath: "generator/templates/responses.go.tmpl",
	ResultPath:   "responses.go",
}

var EndpointTemplate = Template{
	TemplatePath: "generator/templates/endpoints.go.tmpl",
	ResultPath:   "endpoints.go",
}
