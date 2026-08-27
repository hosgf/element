package trace

import (
	"context"
	"strings"

	"github.com/hosgf/element/types"
)

// FirstHint 返回第一个非空字符串。
func FirstHint(vals ...string) string {
	for _, v := range vals {
		if s := strings.TrimSpace(v); s != "" {
			return s
		}
	}
	return ""
}

// Continue 复用已有 trace_id；否则用 hint；再否则生成，并写入 OTEL。
func Continue(ctx context.Context, hint ...string) context.Context {
	ctx = bg(ctx)
	if tid := get(ctx, types.TraceIdKey); tid != "" {
		return withOTEL(ctx, tid)
	}
	tid := FirstHint(hint...)
	if tid == "" {
		tid = newID()
	}
	return put(ctx, tid)
}

// Start 新开 trace_id。
func Start(ctx context.Context) context.Context {
	return put(bg(ctx), newID())
}

func put(ctx context.Context, tid string) context.Context {
	if tid == "" {
		return ctx
	}
	return withOTEL(context.WithValue(ctx, types.TraceIdKey, tid), tid)
}

func bg(ctx context.Context) context.Context {
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
