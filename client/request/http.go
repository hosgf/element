package request

import (
	"context"
	"strings"

	"github.com/hosgf/element/trace"
	"github.com/hosgf/element/types"
)

// RequestID 读取 ctx 中的 request_id（与 ctx.GetReqId 一致）。
func RequestID(c context.Context) string {
	return strings.TrimSpace(get(c, types.RequestIdKey))
}

// EnsureRequestID 仅 HTTP 入站：复用 / hint / 生成 request_id。
func EnsureRequestID(c context.Context, hint ...string) context.Context {
	c = bg(c)
	if RequestID(c) != "" {
		return c
	}
	rid := trace.FirstHint(hint...)
	if rid == "" {
		rid = GenerateRequestID()
	}
	if rid == "" {
		return c
	}
	return context.WithValue(c, types.RequestIdKey, rid)
}

// ApplyHTTP 一次补齐 HTTP 的 trace + request（中间件共用）。
func ApplyHTTP(c context.Context, traceHint, reqHint string) context.Context {
	c = trace.Continue(c, traceHint)
	return EnsureRequestID(c, reqHint)
}

func bg(c context.Context) context.Context {
	if c == nil {
		return context.Background()
	}
	return c
}

func get(c context.Context, key string) string {
	if c == nil {
		return ""
	}
	v, _ := c.Value(key).(string)
	return v
}
