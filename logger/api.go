package logger

import (
	"context"
	"sync"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/glog"
)

// Logger 统一日志接口。
// 级别方法会从 ctx 读取 request_id / trace_id 并附加到输出；
// 业务打日志请用接口方法，配置底层实例请用 Raw()。
type Logger interface {
	// Print 输出不分级日志。
	Print(ctx context.Context, v ...interface{})
	// Printf 按 format 输出不分级日志。
	Printf(ctx context.Context, format string, v ...interface{})
	// Debug 输出 Debug 日志。
	Debug(ctx context.Context, v ...interface{})
	// Debugf 按 format 输出 Debug 日志。
	Debugf(ctx context.Context, format string, v ...interface{})
	// Info 输出 Info 日志。
	Info(ctx context.Context, v ...interface{})
	// Infof 按 format 输出 Info 日志。
	Infof(ctx context.Context, format string, v ...interface{})
	// Notice 输出 Notice 日志。
	Notice(ctx context.Context, v ...interface{})
	// Noticef 按 format 输出 Notice 日志。
	Noticef(ctx context.Context, format string, v ...interface{})
	// Warning 输出 Warning 日志。
	Warning(ctx context.Context, v ...interface{})
	// Warningf 按 format 输出 Warning 日志。
	Warningf(ctx context.Context, format string, v ...interface{})
	// Error 输出 Error 日志。
	Error(ctx context.Context, v ...interface{})
	// Errorf 按 format 输出 Error 日志。
	Errorf(ctx context.Context, format string, v ...interface{})
	// Critical 输出 Critical 日志。
	Critical(ctx context.Context, v ...interface{})
	// Criticalf 按 format 输出 Critical 日志。
	Criticalf(ctx context.Context, format string, v ...interface{})
	// Fatal 输出 Fatal 日志后退出进程。
	Fatal(ctx context.Context, v ...interface{})
	// Fatalf 按 format 输出 Fatal 日志后退出进程。
	Fatalf(ctx context.Context, format string, v ...interface{})
	// Panic 输出 Panic 日志后触发 panic。
	Panic(ctx context.Context, v ...interface{})
	// Panicf 按 format 输出 Panic 日志后触发 panic。
	Panicf(ctx context.Context, format string, v ...interface{})
	// PrintStack 打印当前调用栈。
	PrintStack(ctx context.Context, skip ...int)
	// SetLevelStr 按字符串设置日志级别，如 "all" / "debug" / "info"。
	SetLevelStr(levelStr string) error
	// Raw 返回底层 *glog.Logger，仅用于 Path/Config 等配置，不要直接 Infof。
	Raw() *glog.Logger
}

// wrap 包装 *glog.Logger。
// hook 为 true 时，写日志后触发 AddHandler（默认库双写）；
// 为 false 时仅附加 meta，用于 Named/New 多文件场景，避免 Handler 回环。
type wrap struct {
	raw  *glog.Logger
	hook bool
}

var (
	mu   sync.RWMutex
	base *glog.Logger // 底层默认实例
	std  Logger       // 包装后的默认 Logger
)

// SetDefault 设置包级默认 *glog.Logger，并刷新 Default()。
// 传入 nil 会清空缓存，下次 Default()/Log() 再从 g.Log() 懒加载。
func SetDefault(l *glog.Logger) {
	mu.Lock()
	base = l
	if l != nil {
		std = &wrap{raw: l, hook: true}
	} else {
		std = nil
	}
	mu.Unlock()
}

// Log 返回底层 *glog.Logger。
//   - name 为空：与 Default() 同源，供仍接收 *glog.Logger 的旧 API；
//   - name 非空：返回 g.Log(name) 分类实例（无 meta 附加）。
//
// 需要带 request_id/trace_id 的多文件写入请用 Named / New。
func Log(name ...string) *glog.Logger {
	if len(name) > 0 {
		return g.Log(name...)
	}
	return Default().Raw()
}

// Wrap 包装 *glog.Logger，并启用 AddHandler（用于默认库或需双写的场景）。
func Wrap(raw *glog.Logger) Logger {
	return &wrap{raw: raw, hook: true}
}

// New 包装独立 *glog.Logger（可自定义 Path/Config），附加 meta，不触发 Handler。
// 适用于审计日志、插件日志等独立文件写入。
func New(raw *glog.Logger) Logger {
	return &wrap{raw: raw, hook: false}
}

// Named 按名称取分类 Logger（对应 g.Log(name)，可在配置中单独设文件路径）。
// 附加 meta，不触发 Handler。
func Named(name string) Logger {
	return New(g.Log(name))
}

// Default 返回当前默认 Logger（与 SetDefault / 无参 Log() 同源）。
// 未设置时懒加载 g.Log() 并缓存。
func Default() Logger {
	mu.RLock()
	l := std
	mu.RUnlock()
	if l != nil {
		return l
	}
	mu.Lock()
	defer mu.Unlock()
	if std != nil {
		return std
	}
	if base == nil {
		base = g.Log()
	}
	std = &wrap{raw: base, hook: true}
	return std
}

// Raw 返回底层 *glog.Logger。
func (w *wrap) Raw() *glog.Logger { return w.raw }

// SetLevelStr 设置底层日志级别。
func (w *wrap) SetLevelStr(levelStr string) error {
	return w.raw.SetLevelStr(levelStr)
}

// writef 格式化写入：附加 meta，可选触发 Handler。
func (w *wrap) writef(ctx context.Context, level string, fn func(context.Context, string, ...interface{}), format string, v ...interface{}) {
	ctx = safeCtx(ctx)
	format = Format(ctx, format)
	fn(ctx, format, v...)
	if w.hook {
		emit(ctx, level, format, v)
	}
}

// writev 可变参写入：附加 meta，可选触发 Handler。
func (w *wrap) writev(ctx context.Context, level string, fn func(context.Context, ...interface{}), v ...interface{}) {
	ctx = safeCtx(ctx)
	v = withMeta(ctx, v)
	fn(ctx, v...)
	if w.hook {
		emit(ctx, level, "", v)
	}
}

// Print 实现 Logger。
func (w *wrap) Print(ctx context.Context, v ...interface{}) {
	w.writev(ctx, "print", w.raw.Print, v...)
}

// Printf 实现 Logger。
func (w *wrap) Printf(ctx context.Context, format string, v ...interface{}) {
	w.writef(ctx, "print", w.raw.Printf, format, v...)
}

// Debug 实现 Logger。
func (w *wrap) Debug(ctx context.Context, v ...interface{}) {
	w.writev(ctx, "debug", w.raw.Debug, v...)
}

// Debugf 实现 Logger。
func (w *wrap) Debugf(ctx context.Context, format string, v ...interface{}) {
	w.writef(ctx, "debug", w.raw.Debugf, format, v...)
}

// Info 实现 Logger。
func (w *wrap) Info(ctx context.Context, v ...interface{}) {
	w.writev(ctx, "info", w.raw.Info, v...)
}

// Infof 实现 Logger。
func (w *wrap) Infof(ctx context.Context, format string, v ...interface{}) {
	w.writef(ctx, "info", w.raw.Infof, format, v...)
}

// Notice 实现 Logger。
func (w *wrap) Notice(ctx context.Context, v ...interface{}) {
	w.writev(ctx, "notice", w.raw.Notice, v...)
}

// Noticef 实现 Logger。
func (w *wrap) Noticef(ctx context.Context, format string, v ...interface{}) {
	w.writef(ctx, "notice", w.raw.Noticef, format, v...)
}

// Warning 实现 Logger。
func (w *wrap) Warning(ctx context.Context, v ...interface{}) {
	w.writev(ctx, "warning", w.raw.Warning, v...)
}

// Warningf 实现 Logger。
func (w *wrap) Warningf(ctx context.Context, format string, v ...interface{}) {
	w.writef(ctx, "warning", w.raw.Warningf, format, v...)
}

// Error 实现 Logger。
func (w *wrap) Error(ctx context.Context, v ...interface{}) {
	w.writev(ctx, "error", w.raw.Error, v...)
}

// Errorf 实现 Logger。
func (w *wrap) Errorf(ctx context.Context, format string, v ...interface{}) {
	w.writef(ctx, "error", w.raw.Errorf, format, v...)
}

// Critical 实现 Logger。
func (w *wrap) Critical(ctx context.Context, v ...interface{}) {
	w.writev(ctx, "critical", w.raw.Critical, v...)
}

// Criticalf 实现 Logger。
func (w *wrap) Criticalf(ctx context.Context, format string, v ...interface{}) {
	w.writef(ctx, "critical", w.raw.Criticalf, format, v...)
}

// Fatal 实现 Logger。
func (w *wrap) Fatal(ctx context.Context, v ...interface{}) {
	w.writev(ctx, "fatal", w.raw.Fatal, v...)
}

// Fatalf 实现 Logger。
func (w *wrap) Fatalf(ctx context.Context, format string, v ...interface{}) {
	w.writef(ctx, "fatal", w.raw.Fatalf, format, v...)
}

// Panic 实现 Logger。
func (w *wrap) Panic(ctx context.Context, v ...interface{}) {
	w.writev(ctx, "panic", w.raw.Panic, v...)
}

// Panicf 实现 Logger。
func (w *wrap) Panicf(ctx context.Context, format string, v ...interface{}) {
	w.writef(ctx, "panic", w.raw.Panicf, format, v...)
}

// PrintStack 实现 Logger。
func (w *wrap) PrintStack(ctx context.Context, skip ...int) {
	w.raw.PrintStack(safeCtx(ctx), skip...)
}
