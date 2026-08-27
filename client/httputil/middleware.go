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
	// Propagate 按请求 ctx 覆盖链路头，避免客户端构造时预置的旧 ID 残留
	hs := []gclient.HandlerFunc{MiddlewarePropagate}
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

// MiddlewarePropagate 每请求从 r.Context() 注入 / 覆盖 X-Trace-Id、X-Req-Id。
func MiddlewarePropagate(c *gclient.Client, r *http.Request) (resp *gclient.Response, err error) {
	request.Inject(r.Header, r.Context())
	return c.Next(r)
}

func middlewareHeader(ctx context.Context, c *gclient.Client) *gclient.Client {
	if headers := request.GetHeader(ctx); len(headers) > 0 {
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
