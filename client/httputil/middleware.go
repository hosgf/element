package httputil

import (
	"context"
	"net/http"

	"github.com/gogf/gf/v2/net/gclient"
	"github.com/hosgf/element/client/request"
)

func NewClient(ctx context.Context, client *gclient.Client) *Client {
	c := &Client{
		ctx: ctx,
		c:   client,
	}
	c.SetMiddleware()
	return c
}

type Client struct {
	ctx context.Context
	c   *gclient.Client
}

type MiddlewareFunc = func(ctx context.Context, c *gclient.Client) *gclient.Client

func (c *Client) SetMiddleware(handlers ...gclient.HandlerFunc) *Client {
	c.middleware(middlewareHeader, middlewareCookies)
	hs := []gclient.HandlerFunc{MiddlewareSame, MiddlewareSecurity}
	if len(handlers) > 0 {
		hs = append(hs, handlers...)
	}
	return c.use(hs...)
}

func (c *Client) use(handlers ...gclient.HandlerFunc) *Client {
	if len(handlers) < 1 {
		return c
	}
	c.c.Use(handlers...)
	return c
}

func (c *Client) middleware(middlewares ...MiddlewareFunc) {
	if len(middlewares) < 1 {
		return
	}
	for _, middleware := range middlewares {
		c.c = middleware(c.ctx, c.c)
	}
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

func middlewareCookies(ctx context.Context, c *gclient.Client) *gclient.Client {
	if cookies := request.GetDefaultCookies(ctx); cookies != nil {
		return c.Cookie(cookies)
	}
	return c
}
