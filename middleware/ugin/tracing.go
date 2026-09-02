package ugin

import (
	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
)

// Tracing 创建 Gin SERVER Span 并从 traceparent 提取父上下文。
// 须在 MiddlewareHeader 之前注册：r.Use(ugin.Tracing("service-name"))。
func Tracing(serviceName string) gin.HandlerFunc {
	return otelgin.Middleware(serviceName)
}
