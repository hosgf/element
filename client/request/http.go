package request

import (
	"context"
	"net/http"
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

// Outbound 从 ctx 提取出站链路头（X-Trace-Id / X-Req-Id）。
// 语义 key（trace_id / request_id）优先，其次 header 名。
func Outbound(c context.Context) map[string]string {
	out := make(map[string]string, 2)
	if tid := HeaderTraceId.Get(c); tid != "" {
		out[HeaderTraceId.String()] = tid
	}
	if rid := HeaderReqId.Get(c); rid != "" {
		out[HeaderReqId.String()] = rid
	}
	return out
}

// Inject 用 ctx 中的链路头写入 h（有值则覆盖，保证与当前请求一致）。
func Inject(h http.Header, c context.Context) {
	if h == nil {
		return
	}
	for k, v := range Outbound(c) {
		h.Set(k, v)
	}
}

func firstID(c context.Context, keys ...string) string {
	for _, key := range keys {
		if s := strings.TrimSpace(get(c, key)); s != "" {
			return s
		}
	}
	return ""
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
