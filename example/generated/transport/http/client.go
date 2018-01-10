// This file was automatically generated by "microgen 0.7.0b" utility.
// Please, do not edit.
package transporthttp

import (
	generated "github.com/devimteam/microgen/example/generated"
	http1 "github.com/devimteam/microgen/example/generated/transport/converter/http"
	http "github.com/go-kit/kit/transport/http"
	url "net/url"
	strings "strings"
)

func NewHTTPClient(addr string, opts ...http.ClientOption) (generated.StringService, error) {
	if !strings.HasPrefix(addr, "http") {
		addr = "http://" + addr
	}
	u, err := url.Parse(addr)
	if err != nil {
		return nil, err
	}
	return &generated.Endpoints{
		CountEndpoint: http.NewClient(
			"POST",
			u,
			http1.EncodeHTTPCountRequest,
			http1.DecodeHTTPCountResponse,
			opts...,
		).Endpoint(),
		TestCaseEndpoint: http.NewClient(
			"POST",
			u,
			http1.EncodeHTTPTestCaseRequest,
			http1.DecodeHTTPTestCaseResponse,
			opts...,
		).Endpoint(),
		UppercaseEndpoint: http.NewClient(
			"POST",
			u,
			http1.EncodeHTTPUppercaseRequest,
			http1.DecodeHTTPUppercaseResponse,
			opts...,
		).Endpoint(),
	}, nil
}
