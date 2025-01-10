package request

import (
	"github.com/gin-gonic/gin"
)

func RequestInterceptor() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.GetHeader("")
		ctx := c.Request.Context()
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}
