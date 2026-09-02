package trace

import (
	"context"
	"strings"

	"go.opentelemetry.io/otel/trace"

	"github.com/hosgf/element/types"
)

// First 返回第一个非空字符串。
func First(vals ...string) string {
	for _, v := range vals {
		if s := strings.TrimSpace(v); s != "" {
			return s
		}
	}
	return ""
}

// Continue 复用已有 trace_id；否则用 hint；再否则生成。
// 有效 OTel SpanContext 优先，不再根据自定义头伪造 Remote SpanContext。
func Continue(ctx context.Context, hint ...string) context.Context {
	if ctx = safe(ctx); get(ctx, types.TraceIdKey) != "" {
		return ctx
	}
	if sc := trace.SpanContextFromContext(ctx); sc.IsValid() {
		return set(ctx, sc.TraceID().String())
	}
	if tid := First(hint...); tid != "" {
		return set(ctx, tid)
	}
	return set(ctx, newID())
}

// Start 新开 trace_id（仅写入语义 key，不创建 OTel Span）。
func Start(ctx context.Context) context.Context {
	return set(safe(ctx), newID())
}

func set(ctx context.Context, tid string) context.Context {
	if tid != "" {
		ctx = context.WithValue(ctx, types.TraceIdKey, tid)
	}
	return ctx
}

func safe(ctx context.Context) context.Context {
	if ctx == nil {
		return context.Background()
	}
	return ctx
}

func get(ctx context.Context, key string) string {
	if ctx == nil {
		return ""
	}
	v, _ := ctx.Value(key).(string)
	return strings.TrimSpace(v)
}
