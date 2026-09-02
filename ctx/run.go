package ctx

import (
	"context"

	"go.opentelemetry.io/otel"
	oteltrace "go.opentelemetry.io/otel/trace"

	"github.com/hosgf/element/trace"
)

const instrumentationName = "github.com/hosgf/element/ctx"

// ContinueTrace 复用已有 trace_id；否则用 hint；再否则生成。
func ContinueTrace(ctx context.Context, hint ...string) context.Context {
	return trace.Continue(ctx, hint...)
}

// StartTrace 新开 trace_id；多数场景请直接用 Run / RunNamed。
func StartTrace(ctx context.Context) context.Context {
	return trace.Start(ctx)
}

// Run 新开 Root Span 并同步执行 fn（定时器 / 监听器入口）。
func Run(c context.Context, fn func(context.Context)) {
	RunNamed(c, "ctx.Run", fn)
}

// RunNamed 以 name 创建 Root Span，回调结束后自动 End。
func RunNamed(c context.Context, name string, fn func(context.Context)) {
	if fn == nil {
		return
	}
	tracer := otel.Tracer(instrumentationName)
	runCtx, span := tracer.Start(Context(c), name, oteltrace.WithNewRoot())
	defer span.End()
	fn(trace.Sync(runCtx))
}

// Go 继承 ctx 值，去 cancel，在 goroutine 中执行 fn。
func Go(c context.Context, fn func(context.Context)) {
	if fn == nil {
		return
	}
	go fn(Fork(c))
}

// Fork 继承父 ctx 值并去掉 cancel（Go 内部使用；需手动 go 时再用）。
func Fork(c context.Context) context.Context {
	if c == nil {
		return context.Background()
	}
	return context.WithoutCancel(c)
}
