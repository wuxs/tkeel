// Code generated by protoc-gen-go-http. DO NOT EDIT.
// versions:
// protoc-gen-go-http 0.1.0

package v1

import (
	go_restful "github.com/emicklei/go-restful"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the tkeel package it is being compiled against.
// import package.context.http.anypb.result.protojson.go_restful.errors.emptypb.

type MetricsHTTPHandler interface {
	Metrics(req *go_restful.Request, resp *go_restful.Response)
}

func RegisterMetricsHTTPServer(container *go_restful.Container, metricHandler MetricsHTTPHandler) {
	var ws *go_restful.WebService
	for _, v := range container.RegisteredWebServices() {
		if v.RootPath() == "/v1" {
			ws = v
			break
		}
	}
	if ws == nil {
		ws = new(go_restful.WebService)
		ws.ApiVersion("/v1")
		ws.Path("/v1")
		container.Add(ws)
	}

	ws.Route(ws.GET("/metrics").
		To(metricHandler.Metrics))
}
