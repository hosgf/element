package goframe

import (
	"strings"

	"github.com/gogf/gf/v2/net/ghttp"

	"github.com/hosgf/element/client/request"
	"github.com/hosgf/element/types"
)

type headerBinding struct {
	ctxKey string
	header request.Header
	gen    func() string
}

var contextHeaders = []headerBinding{
	{types.TraceIdKey, request.HeaderTraceId, request.GenerateTraceID},
	{types.RequestIdKey, request.HeaderReqId, request.GenerateRequestID},
	{types.TenantIdKey, request.HeaderTenantId, nil},
	{types.UserIdKey, request.HeaderUserId, nil},
	{types.ClientIPKey, request.HeaderRealIP, nil},
	{types.DeviceCodeKey, request.HeaderDeviceCode, nil},
	{types.DeviceTypeKey, request.HeaderDeviceType, nil},
	{types.UserAgentKey, request.HeaderUserAgent, nil},
	{types.ReqClientKey, request.HeaderReqClient, nil},
}

func SetMiddleware(s *ghttp.Server, handlers ...ghttp.HandlerFunc) *ghttp.Server {
	hs := make([]ghttp.HandlerFunc, 0, 3+len(handlers))
	hs = append(hs, MiddlewareCORS, MiddlewareHeader, MiddlewareCookies)
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
	for _, b := range contextHeaders {
		WithValue(r, b.ctxKey, b.header, b.gen)
	}
}

func SetHeaders(r *ghttp.Request, headers ...request.Header) {
	for _, header := range headers {
		SetHandler(r, header)
	}
}

func MiddlewareCookies(r *ghttp.Request) {
	SetCookies(r)
	r.Middleware.Next()
}

func SetCookies(req *ghttp.Request) *ghttp.Request {
	cookies := req.Cookies()
	if len(cookies) == 0 {
		return req
	}
	cookieMap := make(map[string]string, len(cookies))
	for _, c := range cookies {
		cookieMap[c.Name] = c.Value
	}
	req.SetCtxVar(request.CookieKey, cookieMap)
	return req
}

// SetHandler 将非空请求头写入同名 ctx 变量，用于后续 HTTP 透传。
func SetHandler(req *ghttp.Request, header request.Header) *ghttp.Request {
	if value := GetHeader(req, header); value != "" {
		req.SetCtxVar(header.String(), value)
	}
	return req
}

func GetHeader(req *ghttp.Request, key request.Header) string {
	return strings.TrimSpace(req.GetHeader(key.String()))
}

// WithValue 优先用请求头；无则用 gen 生成；写入语义 ctxKey 与 header 名。
func WithValue(req *ghttp.Request, ctxKey string, header request.Header, gen func() string) *ghttp.Request {
	value := GetHeader(req, header)
	if value == "" {
		if gen == nil {
			return req
		}
		value = gen()
		if value == "" {
			return req
		}
		req.Header.Set(header.String(), value)
	}
	setCtx(req, ctxKey, value)
	setCtx(req, header.String(), value)
	return req
}

func setCtx(req *ghttp.Request, key, value string) {
	if req.GetCtxVar(key).String() != "" {
		return
	}
	req.SetCtxVar(key, value)
}
