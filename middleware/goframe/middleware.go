package goframe

import (
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/hosgf/element/client/request"
)

func SetMiddleware(s *ghttp.Server, handlers ...ghttp.HandlerFunc) *ghttp.Server {
	hs := []ghttp.HandlerFunc{MiddlewareCORS, MiddlewareHeader, MiddlewareCookies}
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
	for _, header := range request.GetHeaders() {
		r = SetHandler(r, header)
	}
	r.Middleware.Next()
}

func MiddlewareCookies(r *ghttp.Request) {
	r = SetCookies(r)
	r.Middleware.Next()
}

func SetCookies(req *ghttp.Request) *ghttp.Request {
	cookies := req.Cookies()
	cookieMap := make(map[string]string)
	for _, cookie := range cookies {
		cookieMap[cookie.Name] = cookie.Value
	}
	if len(cookieMap) < 1 {
		return req
	}
	req.SetCtxVar(request.CookieKey, cookieMap)
	return req
}

func SetHandler(req *ghttp.Request, header request.Header) *ghttp.Request {
	if value := GetHeader(req, header); len(value) > 0 {
		req.SetCtxVar(header.String(), value)
	}
	return req
}

func GetHeader(req *ghttp.Request, key request.Header) string {
	return req.GetHeader(key.String())
}
