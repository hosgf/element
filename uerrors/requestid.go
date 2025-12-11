package uerrors

import (
	"crypto/rand"
	"encoding/hex"
	"time"
)

// GenerateRequestID 生成统一的请求ID
func GenerateRequestID() string {
	buf := make([]byte, 12)
	_, _ = rand.Read(buf)
	return time.Now().Format("20060102150405") + "-" + hex.EncodeToString(buf)
}
