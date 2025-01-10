package request

import (
	"context"
	"github.com/gin-gonic/gin"
)

func RequestInterceptor() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		c.GetAppCode("")
		ctx = context.WithValue(ctx, HeaderKey, headers)
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}
