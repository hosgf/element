package request

import (
	"context"
	"net/http"
	"strings"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"

	"github.com/hosgf/element/trace"
	"github.com/hosgf/element/types"
)

// ID 读取 ctx 中的 request_id（与 ctx.GetReqId 一致）。
func ID(c context.Context) string {
	return strings.TrimSpace(get(c, types.RequestIdKey))
}

// EnsureID 仅 HTTP 入站：复用 / hint / 生成 request_id。
func EnsureID(c context.Context, hint ...string) context.Context {
	c = safe(c)
	if ID(c) != "" {
		return c
	}
	rid := trace.First(hint...)
	if rid == "" {
		rid = GenerateRequestID()
	}
	if rid == "" {
		return c
	}
	return context.WithValue(c, types.RequestIdKey, rid)
}

// ApplyHTTP 补齐 HTTP request_id；trace 由 OTel traceparent 传播，traceHint 已忽略。
func ApplyHTTP(c context.Context, traceHint, reqHint string) context.Context {
	_ = traceHint
	return EnsureID(c, reqHint)
}

// Outbound 生成出站 HTTP 头：traceparent（OTel inject）+ 新 X-Req-Id。
func Outbound(c context.Context) map[string]string {
	h := make(http.Header)
	Inject(h, c)
	out := make(map[string]string)
	for _, field := range otel.GetTextMapPropagator().Fields() {
		if v := h.Get(field); v != "" {
			out[field] = v
		}
	}
	if rid := h.Get(HeaderReqId.String()); rid != "" {
		out[HeaderReqId.String()] = rid
	}
	return out
}

// Inject 注入 traceparent 与全新 X-Req-Id（每次出站一个新 ID，供 Dedup）。
func Inject(h http.Header, c context.Context) {
	if h == nil {
		return
	}
	otel.GetTextMapPropagator().Inject(c, propagation.HeaderCarrier(h))
	h.Set(HeaderReqId.String(), GenerateRequestID())
}

func firstID(c context.Context, keys ...string) string {
	for _, key := range keys {
		if s := strings.TrimSpace(get(c, key)); s != "" {
			return s
		}
	}
	return ""
}

func safe(c context.Context) context.Context {
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
