package ugin

import (
	"context"
	"strings"

	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
	"github.com/hosgf/element/client/request"
	"github.com/hosgf/element/middleware"
	"github.com/hosgf/element/trace"
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
	return trace.Sync(c.Request.Context())
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
