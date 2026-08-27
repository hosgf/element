package ctx

import (
	"context"

	"github.com/hosgf/element/trace"
)

// ContinueTrace 复用已有 trace_id；否则用 hint；再否则生成，并写入 OTEL。
func ContinueTrace(ctx context.Context, hint ...string) context.Context {
	return trace.Continue(ctx, hint...)
}

// StartTrace 新开 trace_id；多数场景请直接用 Run。
func StartTrace(ctx context.Context) context.Context {
	return trace.Start(ctx)
}

// Run 新开 trace 并同步执行 fn（定时器 / 监听器入口）。
func Run(c context.Context, fn func(context.Context)) {
	if fn == nil {
		return
	}
	fn(StartTrace(c))
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
