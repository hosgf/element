package goframe

import (
	"bytes"
	"compress/gzip"
	"strings"

	"github.com/gogf/gf/v2/net/ghttp"
)

// MiddlewareGzip 在 handler 之后按 Accept-Encoding 压缩 buffer。
// 须作为外层中间件（SetMiddleware 已放在最前），以便内侧 AccessLog 读到明文 JSON。
func MiddlewareGzip(r *ghttp.Request) {
	r.Middleware.Next()

	if r.Response.Header().Get("Content-Encoding") != "" {
		return
	}
	if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
		return
	}
	buf := r.Response.Buffer()
	if len(buf) == 0 {
		return
	}

	var compressed bytes.Buffer
	gw := gzip.NewWriter(&compressed)
	if _, err := gw.Write(buf); err != nil {
		return
	}
	if err := gw.Close(); err != nil {
		return
	}

	r.Response.ClearBuffer()
	r.Response.Header().Set("Content-Encoding", "gzip")
	r.Response.Header().Del("Content-Length")
	r.Response.Write(compressed.Bytes())
}
