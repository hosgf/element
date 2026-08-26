package request

import (
	"crypto/rand"
	"encoding/hex"
	"time"
)

// GenerateRequestID 业务请求 ID（幂等/去重），不是 TraceId。
func GenerateRequestID() string {
	buf := make([]byte, 12)
	if _, err := rand.Read(buf); err != nil {
		return ""
	}
	return time.Now().Format("20060102150405") + hex.EncodeToString(buf)
}

// GenerateTraceID 生成 32 hex TraceId（OTEL 形状），经 X-Trace-Id 传递。
func GenerateTraceID() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return ""
	}
	return hex.EncodeToString(b)
}
