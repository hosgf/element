package request

import (
	"crypto/rand"
	"encoding/hex"
	"strconv"
	"time"
)

// GenerateRequestID 生成统一的请求ID
func GenerateRequestID() string {
	buf := make([]byte, 12)
	if _, err := rand.Read(buf); err != nil {
		return time.Now().Format("20060102150405") + strconv.FormatInt(time.Now().UnixNano(), 10)
	}
	return time.Now().Format("20060102150405") + hex.EncodeToString(buf)
}
