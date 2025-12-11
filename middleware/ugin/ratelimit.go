package ugin

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"

	"github.com/hosgf/element/model/result"
)

// RateLimiter 创建限流中间件
// rateLimit: 每秒允许的请求数
// burst: 突发请求的最大数量
func RateLimiter(rateLimit float64, burst int) gin.HandlerFunc {
	limiter := rate.NewLimiter(rate.Limit(rateLimit), burst)
	return func(c *gin.Context) {
		if limiter.Allow() {
			c.Next()
		} else {
			response := result.NewResponse()
			response.Code = http.StatusTooManyRequests
			response.Message = "请求过于频繁，请稍后再试"
			c.Status(http.StatusOK)
			c.JSON(http.StatusOK, response)
			c.Abort()
		}
	}
}

// RateLimiterDefault 使用默认参数的限流中间件
// 默认: rate=10, burst=30
func RateLimiterDefault() gin.HandlerFunc {
	return RateLimiter(10, 30)
}
