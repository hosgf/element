package request

import (
	"crypto/rand"
	"encoding/hex"
	"time"
)

// GenerateRequestID 生成业务请求 ID（幂等/去重），不是 trace_id。
func GenerateRequestID() string {
	buf := make([]byte, 12)
	if _, err := rand.Read(buf); err != nil {
		return ""
	}
	return time.Now().Format("20060102150405") + hex.EncodeToString(buf)
}
