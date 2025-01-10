package request

import (
	"context"

	"github.com/gin-gonic/gin"
)

func RequestInterceptor() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		for _, header := range GetHeaders() {
			ctx = setHandler(ctx, c, header)
		}
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}

func setHandler(ctx context.Context, c *gin.Context, header Header) context.Context {
	if value := GetHeader(c, header); len(value) > 0 {
		ctx = context.WithValue(ctx, header, value)
		c.Set(header.String(), value)
	}
	return ctx
}
