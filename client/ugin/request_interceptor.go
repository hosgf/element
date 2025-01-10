package ugin

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/hosgf/element/client/request"
)

func RequestInterceptor() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		for _, header := range request.GetHeaders() {
			ctx = setHandler(ctx, c, header)
		}
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}

func setHandler(ctx context.Context, c *gin.Context, header request.Header) context.Context {
	if value := request.GetHeader(c, header); len(value) > 0 {
		ctx = context.WithValue(ctx, header, value)
		c.Set(header.String(), value)
	}
	return ctx
}
