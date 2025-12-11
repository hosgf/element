package goframe

import (
	"net/http"

	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/hosgf/element/model/result"
	"golang.org/x/time/rate"
)

// RateLimiter 创建限流中间件
// rateLimit: 每秒允许的请求数
// burst: 突发请求的最大数量
func RateLimiter(rateLimit float64, burst int) ghttp.HandlerFunc {
	limiter := rate.NewLimiter(rate.Limit(rateLimit), burst)
	return func(r *ghttp.Request) {
		if limiter.Allow() {
			r.Middleware.Next()
		} else {
			response := result.NewResponse()
			response.Code = http.StatusTooManyRequests
			response.Message = "请求过于频繁，请稍后再试"
			result.Writer(r, response)
		}
	}
}

// RateLimiterDefault 使用默认参数的限流中间件
// 默认: rate=10, burst=30
func RateLimiterDefault() ghttp.HandlerFunc {
	return RateLimiter(10, 30)
}
