package httputil

import (
	"context"
	"net/http"

	"github.com/gogf/gf/v2/net/gclient"
	"github.com/hosgf/element/client/request"
)

func SetMiddleware(ctx context.Context, c *gclient.Client, handlers ...gclient.HandlerFunc) *gclient.Client {
	c = middlewareHeader(ctx, c)
	hs := []gclient.HandlerFunc{MiddlewareSame, MiddlewareSecurity}
	if len(handlers) > 0 {
		hs = append(hs, handlers...)
	}
	c.Use(hs...)
	return c
}

func MiddlewareSame(c *gclient.Client, r *http.Request) (resp *gclient.Response, err error) {
	return c.Next(r)
}

func MiddlewareSecurity(c *gclient.Client, r *http.Request) (resp *gclient.Response, err error) {
	return c.Next(r)
}

func middlewareHeader(ctx context.Context, c *gclient.Client) *gclient.Client {
	if headers := request.GetHeader(ctx); headers != nil {
		return c.Header(headers)
	}
	return c
}
