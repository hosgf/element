package ugin

import (
	"context"
	"strings"

	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
	"github.com/hosgf/element/client/request"
	"github.com/hosgf/element/ctx"
	"github.com/hosgf/element/middleware"
	"github.com/hosgf/element/types"
)

func SetMiddleware(s *gin.Engine, handlers ...gin.HandlerFunc) *gin.Engine {
	hs := []gin.HandlerFunc{gzip.Gzip(gzip.DefaultCompression), MiddlewareHeader()}
	if len(handlers) > 0 {
		hs = append(hs, handlers...)
	}
	s.Use(hs...)
	return s
}

func MiddlewareHeader() gin.HandlerFunc {
	return func(c *gin.Context) {
		reqCtx := bindTrace(c)
		for _, b := range middleware.ContextHeaders {
			reqCtx = bindHeader(reqCtx, c, b.Key, b.Header)
		}
		for _, header := range request.GetHeaders() {
			reqCtx = passthroughHeader(reqCtx, c, header)
		}
		c.Request = c.Request.WithContext(reqCtx)
		c.Next()
	}
}

func bindTrace(c *gin.Context) context.Context {
	hint := GetHeader(c, request.HeaderTraceId)
	reqCtx := ctx.ContinueTrace(c.Request.Context(), hint)
	tid := ctx.GetTraceId(reqCtx)
	if tid == "" {
		return reqCtx
	}
	// 语义 key 已在 ContinueTrace 写入；补 header 名供出站透传。
	reqCtx = context.WithValue(reqCtx, request.HeaderTraceId.String(), tid)
	c.Set(types.TraceIdKey, tid)
	c.Set(request.HeaderTraceId.String(), tid)
	if hint == "" {
		c.Request.Header.Set(request.HeaderTraceId.String(), tid)
	}
	return reqCtx
}

func bindHeader(reqCtx context.Context, c *gin.Context, key string, header request.Header) context.Context {
	value := GetHeader(c, header)
	if value == "" {
		return reqCtx
	}
	return bindID(reqCtx, c, key, header, value)
}

func passthroughHeader(reqCtx context.Context, c *gin.Context, header request.Header) context.Context {
	if value := GetHeader(c, header); value != "" {
		reqCtx = context.WithValue(reqCtx, header.String(), value)
		c.Set(header.String(), value)
	}
	return reqCtx
}

// bindID 写入语义 key、header 名到 ctx 与 gin。
func bindID(reqCtx context.Context, c *gin.Context, key string, header request.Header, value string) context.Context {
	reqCtx = context.WithValue(reqCtx, key, value)
	reqCtx = context.WithValue(reqCtx, header.String(), value)
	c.Set(key, value)
	c.Set(header.String(), value)
	return reqCtx
}

func GetHeader(c *gin.Context, key request.Header) string {
	return strings.TrimSpace(c.GetHeader(key.String()))
}
