package goframe

import (
	"strings"

	"github.com/gogf/gf/v2/net/ghttp"

	"github.com/hosgf/element/client/request"
	"github.com/hosgf/element/middleware"
	"github.com/hosgf/element/trace"
)

func SetMiddleware(s *ghttp.Server, handlers ...ghttp.HandlerFunc) *ghttp.Server {
	hs := make([]ghttp.HandlerFunc, 0, 4+len(handlers))
	hs = append(hs, MiddlewareGzip, MiddlewareCORS, MiddlewareHeader, MiddlewareCookies)
	hs = append(hs, handlers...)
	s.Use(hs...)
	return s
}

func MiddlewareCORS(r *ghttp.Request) {
	r.Response.CORSDefault()
	r.Middleware.Next()
}

func MiddlewareHeader(r *ghttp.Request) {
	BindContextHeaders(r)
	SetHeaders(r, request.GetHeaders()...)
	r.Middleware.Next()
}

func BindContextHeaders(r *ghttp.Request) {
	bindTrace(r)
	for _, b := range middleware.ContextHeaders {
		bindHeader(r, b.Key, b.Header)
	}
}

func bindTrace(r *ghttp.Request) {
	r.SetCtx(trace.Sync(r.Context()))
}

func SetHeaders(r *ghttp.Request, headers ...request.Header) {
	for _, header := range headers {
		SetHeader(r, header)
	}
}

func MiddlewareCookies(r *ghttp.Request) {
	SetCookies(r)
	r.Middleware.Next()
}

func SetCookies(r *ghttp.Request) *ghttp.Request {
	cookies := r.Cookies()
	if len(cookies) == 0 {
		return r
	}
	cookieMap := make(map[string]string, len(cookies))
	for _, c := range cookies {
		cookieMap[c.Name] = c.Value
	}
	r.SetCtxVar(request.CookieKey, cookieMap)
	return r
}

// SetHeader 将非空请求头写入同名 ctx 变量，用于后续 HTTP 透传。
func SetHeader(r *ghttp.Request, header request.Header) *ghttp.Request {
	if value := GetHeader(r, header); value != "" {
		r.SetCtxVar(header.String(), value)
	}
	return r
}

func GetHeader(r *ghttp.Request, key request.Header) string {
	return strings.TrimSpace(r.GetHeader(key.String()))
}

func bindHeader(r *ghttp.Request, key string, header request.Header) *ghttp.Request {
	value := GetHeader(r, header)
	if value == "" {
		return r
	}
	setCtx(r, key, value)
	setCtx(r, header.String(), value)
	return r
}

func setCtx(r *ghttp.Request, key, value string) {
	if r.GetCtxVar(key).String() != "" {
		return
	}
	r.SetCtxVar(key, value)
}
