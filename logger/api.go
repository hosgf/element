package logger

import (
	"context"
	"sync"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/glog"
)

// Logger 统一日志接口。
// 输出头部由 MetaHandler 处理：{trace_id} <request_id>（有 request 才带 <>）；
// 业务打日志请用接口方法，配置底层实例请用 Raw()。
type Logger interface {
	Print(ctx context.Context, v ...interface{})
	Printf(ctx context.Context, format string, v ...interface{})
	Debug(ctx context.Context, v ...interface{})
	Debugf(ctx context.Context, format string, v ...interface{})
	Info(ctx context.Context, v ...interface{})
	Infof(ctx context.Context, format string, v ...interface{})
	Notice(ctx context.Context, v ...interface{})
	Noticef(ctx context.Context, format string, v ...interface{})
	Warning(ctx context.Context, v ...interface{})
	Warningf(ctx context.Context, format string, v ...interface{})
	Error(ctx context.Context, v ...interface{})
	Errorf(ctx context.Context, format string, v ...interface{})
	Critical(ctx context.Context, v ...interface{})
	Criticalf(ctx context.Context, format string, v ...interface{})
	Fatal(ctx context.Context, v ...interface{})
	Fatalf(ctx context.Context, format string, v ...interface{})
	Panic(ctx context.Context, v ...interface{})
	Panicf(ctx context.Context, format string, v ...interface{})
	PrintStack(ctx context.Context, skip ...int)
	SetLevelStr(levelStr string) error
	Raw() *glog.Logger
}

// wrap 包装 *glog.Logger。
// hook 为 true 时写日志后触发 AddHandler；false 用于 Named/New，避免回环。
type wrap struct {
	raw  *glog.Logger
	hook bool
}

var (
	mu   sync.RWMutex
	base *glog.Logger
	std  Logger
)

// SetDefault 设置包级默认 *glog.Logger，并刷新 Default()。
func SetDefault(l *glog.Logger) {
	mu.Lock()
	base = bindMeta(l)
	if l != nil {
		std = &wrap{raw: base, hook: true}
	} else {
		std = nil
	}
	mu.Unlock()
}

// Log 返回底层 *glog.Logger。
// name 为空与 Default() 同源；非空为 g.Log(name)（已挂 MetaHandler）。
func Log(name ...string) *glog.Logger {
	if len(name) > 0 {
		return bindMeta(g.Log(name...))
	}
	return Default().Raw()
}

// Wrap 包装并启用 AddHandler。
func Wrap(raw *glog.Logger) Logger {
	return &wrap{raw: bindMeta(raw), hook: true}
}

// New 包装独立 logger，不触发 AddHandler。
func New(raw *glog.Logger) Logger {
	return &wrap{raw: bindMeta(raw), hook: false}
}

// Named 按名称取分类 Logger。
func Named(name string) Logger {
	return New(g.Log(name))
}

// Default 返回默认 Logger。
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
		base = bindMeta(g.Log())
	} else {
		base = bindMeta(base)
	}
	std = &wrap{raw: base, hook: true}
	return std
}

func (w *wrap) Raw() *glog.Logger { return w.raw }

func (w *wrap) SetLevelStr(levelStr string) error {
	return w.raw.SetLevelStr(levelStr)
}

func (w *wrap) writef(ctx context.Context, level string, fn func(context.Context, string, ...interface{}), format string, v ...interface{}) {
	ctx = safeCtx(ctx)
	fn(ctx, format, v...)
	if w.hook {
		emit(ctx, level, format, v)
	}
}

func (w *wrap) writev(ctx context.Context, level string, fn func(context.Context, ...interface{}), v ...interface{}) {
	ctx = safeCtx(ctx)
	fn(ctx, v...)
	if w.hook {
		emit(ctx, level, "", v)
	}
}

func (w *wrap) Print(ctx context.Context, v ...interface{}) {
	w.writev(ctx, "print", w.raw.Print, v...)
}
func (w *wrap) Printf(ctx context.Context, format string, v ...interface{}) {
	w.writef(ctx, "print", w.raw.Printf, format, v...)
}
func (w *wrap) Debug(ctx context.Context, v ...interface{}) {
	w.writev(ctx, "debug", w.raw.Debug, v...)
}
func (w *wrap) Debugf(ctx context.Context, format string, v ...interface{}) {
	w.writef(ctx, "debug", w.raw.Debugf, format, v...)
}
func (w *wrap) Info(ctx context.Context, v ...interface{}) {
	w.writev(ctx, "info", w.raw.Info, v...)
}
func (w *wrap) Infof(ctx context.Context, format string, v ...interface{}) {
	w.writef(ctx, "info", w.raw.Infof, format, v...)
}
func (w *wrap) Notice(ctx context.Context, v ...interface{}) {
	w.writev(ctx, "notice", w.raw.Notice, v...)
}
func (w *wrap) Noticef(ctx context.Context, format string, v ...interface{}) {
	w.writef(ctx, "notice", w.raw.Noticef, format, v...)
}
func (w *wrap) Warning(ctx context.Context, v ...interface{}) {
	w.writev(ctx, "warning", w.raw.Warning, v...)
}
func (w *wrap) Warningf(ctx context.Context, format string, v ...interface{}) {
	w.writef(ctx, "warning", w.raw.Warningf, format, v...)
}
func (w *wrap) Error(ctx context.Context, v ...interface{}) {
	w.writev(ctx, "error", w.raw.Error, v...)
}
func (w *wrap) Errorf(ctx context.Context, format string, v ...interface{}) {
	w.writef(ctx, "error", w.raw.Errorf, format, v...)
}
func (w *wrap) Critical(ctx context.Context, v ...interface{}) {
	w.writev(ctx, "critical", w.raw.Critical, v...)
}
func (w *wrap) Criticalf(ctx context.Context, format string, v ...interface{}) {
	w.writef(ctx, "critical", w.raw.Criticalf, format, v...)
}
func (w *wrap) Fatal(ctx context.Context, v ...interface{}) {
	w.writev(ctx, "fatal", w.raw.Fatal, v...)
}
func (w *wrap) Fatalf(ctx context.Context, format string, v ...interface{}) {
	w.writef(ctx, "fatal", w.raw.Fatalf, format, v...)
}
func (w *wrap) Panic(ctx context.Context, v ...interface{}) {
	w.writev(ctx, "panic", w.raw.Panic, v...)
}
func (w *wrap) Panicf(ctx context.Context, format string, v ...interface{}) {
	w.writef(ctx, "panic", w.raw.Panicf, format, v...)
}
func (w *wrap) PrintStack(ctx context.Context, skip ...int) {
	w.raw.PrintStack(safeCtx(ctx), skip...)
}
