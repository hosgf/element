package ugin

import (
	"context"
	"strings"

	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
	"github.com/hosgf/element/client/request"
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
		ctx := c.Request.Context()
		for _, header := range request.GetHeaders() {
			ctx = SetHandler(ctx, c, header)
		}
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}

func SetHandler(ctx context.Context, c *gin.Context, header request.Header) context.Context {
	if value := GetHeader(c, header); value != "" {
		ctx = context.WithValue(ctx, header.String(), value)
		c.Set(header.String(), value)
	}
	return ctx
}

func GetHeader(c *gin.Context, key request.Header) string {
	return strings.TrimSpace(c.GetHeader(key.String()))
}
