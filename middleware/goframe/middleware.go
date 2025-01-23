package goframe

import (
	"context"

	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/os/gctx"
	"github.com/hosgf/element/client/request"
)

func SetMiddleware(s *ghttp.Server, handlers ...ghttp.HandlerFunc) *ghttp.Server {
	hs := []ghttp.HandlerFunc{MiddlewareCORS, MiddlewareHeader}
	if len(handlers) > 0 {
		hs = append(hs, handlers...)
	}
	s.Use(hs...)
	return s
}

func MiddlewareCORS(r *ghttp.Request) {
	r.Response.CORSDefault()
	r.Middleware.Next()
}

func MiddlewareHeader(r *ghttp.Request) {
	ctx := gctx.New()
	for _, header := range request.GetHeaders() {
		ctx = SetHandler(ctx, r, header)
	}
	r.SetCtx(ctx)
	r.Context()
	r.Middleware.Next()
}

func SetHandler(ctx context.Context, req *ghttp.Request, header request.Header) context.Context {
	if value := GetHeader(req, header); len(value) > 0 {
		ctx = context.WithValue(ctx, header, value)
	}
	return ctx
}

func GetHeader(req *ghttp.Request, key request.Header) string {
	return req.GetHeader(key.String())
}
