package goframe

import (
	"github.com/gogf/gf/v2/net/ghttp"

	"github.com/hosgf/element/client/request"
	"github.com/hosgf/element/types"
)

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
	WithValue(r, types.TraceIdKey, request.HeaderTraceId, request.GenerateRequestID)
	WithValue(r, types.RequestIdKey, request.HeaderReqId, request.GenerateRequestID)
	WithValue(r, types.TenantIdKey, request.HeaderTenantId, nil)
	WithValue(r, types.UserIdKey, request.HeaderUserId, nil)
	for _, header := range request.GetHeaders() {
		if header == request.HeaderTraceId ||
			header == request.HeaderReqId ||
			header == request.HeaderTenantId ||
			header == request.HeaderUserId {
			continue
		}
		SetHandler(r, header)
	}
	r.Middleware.Next()
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

// SetHandler 将非空请求头写入同名 ctx 变量（Trace / ReqId 已由 WithValue 写入 types.*，循环中会跳过）。
func SetHandler(req *ghttp.Request, header request.Header) *ghttp.Request {
	if value := GetHeader(req, header); len(value) > 0 {
		req.SetCtxVar(header.String(), value)
	}
	return req
}

func GetHeader(req *ghttp.Request, key request.Header) string {
	return req.GetHeader(key.String())
}

// WithValue 优先用请求头；无请求头时用 defaultID 生成；Ctx 仅写入 ctxKey（types.*），HTTP 仍用 header 名。
func WithValue(req *ghttp.Request, ctxKey string, header request.Header, defaultID func() string) *ghttp.Request {
	if value := GetHeader(req, header); len(value) > 0 {
		SetCtxVar(req, ctxKey, value)
		return req
	}
	if defaultID == nil {
		return req
	}
	val := defaultID()
	if len(val) == 0 {
		return req
	}
	req.Header.Set(header.String(), val)
	req.SetCtxVar(ctxKey, val)
	return req
}

func SetCtxVar(req *ghttp.Request, ctxKey string, value string) {
	if req.GetCtxVar(ctxKey).String() != "" {
		return
	}
	req.SetCtxVar(ctxKey, value)
}
