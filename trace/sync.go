package trace

import (
	"context"

	"github.com/gogf/gf/v2/net/gtrace"

	"github.com/hosgf/element/types"
)

// Sync 将 OTel trace/span ID 同步到 ctx key，供 logger 读取。
func Sync(ctx context.Context) context.Context {
	if ctx == nil {
		return context.Background()
	}
	if tid := gtrace.GetTraceID(ctx); tid != "" {
		ctx = context.WithValue(ctx, types.TraceIdKey, tid)
	}
	if sid := gtrace.GetSpanID(ctx); sid != "" {
		ctx = context.WithValue(ctx, types.SpanIdKey, sid)
	}
	return ctx
}
