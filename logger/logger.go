package logger

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/text/gstr"
)

var (
	handlersMu sync.RWMutex
	handlers   []Handler
)

// Handler 默认日志写出后的额外回调（如插件文件双写）。
// 同 ctx 内再调默认日志不会重入 emit；插件侧请用 Named/New，避免回环。
type Handler func(ctx context.Context, level string, format string, args []interface{})

// AddHandler 注册额外写入钩子，可多次调用；nil 会被忽略。
func AddHandler(h Handler) {
	if h == nil {
		return
	}
	handlersMu.Lock()
	handlers = append(handlers, h)
	handlersMu.Unlock()
}

// SetLevelStr 设置默认 Logger 的日志级别。
func SetLevelStr(levelStr string) error {
	return Default().SetLevelStr(levelStr)
}

// safeCtx 将 nil ctx 替换为 Background，供底层 glog 使用。
func safeCtx(ctx context.Context) context.Context {
	if ctx == nil {
		return context.Background()
	}
	return ctx
}

// emitKey 标记 emit 重入，避免 Handler 内再打默认日志死循环。
type emitKey struct{}

// emit 调用所有已注册 Handler；同 ctx 重入直接返回。
func emit(ctx context.Context, level string, format string, args []interface{}) {
	if ctx != nil {
		if ctx.Value(emitKey{}) != nil {
			return
		}
		ctx = context.WithValue(ctx, emitKey{}, struct{}{})
	}
	handlersMu.RLock()
	hs := handlers
	handlersMu.RUnlock()
	for _, h := range hs {
		h(ctx, level, format, args)
	}
}

// ---- 包级快捷方式：全部委托 Default() ----

// Print 使用默认 Logger 输出不分级日志。
func Print(ctx context.Context, v ...interface{}) { Default().Print(ctx, v...) }

// Printf 使用默认 Logger 按 format 输出不分级日志。
func Printf(ctx context.Context, format string, v ...interface{}) {
	Default().Printf(ctx, format, v...)
}

// Debug 使用默认 Logger 输出 Debug 日志。
func Debug(ctx context.Context, v ...interface{}) { Default().Debug(ctx, v...) }

// Debugf 使用默认 Logger 按 format 输出 Debug 日志。
func Debugf(ctx context.Context, format string, v ...interface{}) {
	Default().Debugf(ctx, format, v...)
}

// Info 使用默认 Logger 输出 Info 日志。
func Info(ctx context.Context, v ...interface{}) { Default().Info(ctx, v...) }

// Infof 使用默认 Logger 按 format 输出 Info 日志。
func Infof(ctx context.Context, format string, v ...interface{}) {
	Default().Infof(ctx, format, v...)
}

// Notice 使用默认 Logger 输出 Notice 日志。
func Notice(ctx context.Context, v ...interface{}) { Default().Notice(ctx, v...) }

// Noticef 使用默认 Logger 按 format 输出 Notice 日志。
func Noticef(ctx context.Context, format string, v ...interface{}) {
	Default().Noticef(ctx, format, v...)
}

// Warning 使用默认 Logger 输出 Warning 日志。
func Warning(ctx context.Context, v ...interface{}) { Default().Warning(ctx, v...) }

// Warningf 使用默认 Logger 按 format 输出 Warning 日志。
func Warningf(ctx context.Context, format string, v ...interface{}) {
	Default().Warningf(ctx, format, v...)
}

// Error 使用默认 Logger 输出 Error 日志。
func Error(ctx context.Context, v ...interface{}) { Default().Error(ctx, v...) }

// Errorf 使用默认 Logger 按 format 输出 Error 日志。
func Errorf(ctx context.Context, format string, v ...interface{}) {
	Default().Errorf(ctx, format, v...)
}

// Critical 使用默认 Logger 输出 Critical 日志。
func Critical(ctx context.Context, v ...interface{}) { Default().Critical(ctx, v...) }

// Criticalf 使用默认 Logger 按 format 输出 Critical 日志。
func Criticalf(ctx context.Context, format string, v ...interface{}) {
	Default().Criticalf(ctx, format, v...)
}

// Fatal 使用默认 Logger 输出 Fatal 日志后退出进程。
func Fatal(ctx context.Context, v ...interface{}) { Default().Fatal(ctx, v...) }

// Fatalf 使用默认 Logger 按 format 输出 Fatal 日志后退出进程。
func Fatalf(ctx context.Context, format string, v ...interface{}) {
	Default().Fatalf(ctx, format, v...)
}

// Panic 使用默认 Logger 输出 Panic 日志后触发 panic。
func Panic(ctx context.Context, v ...interface{}) { Default().Panic(ctx, v...) }

// Panicf 使用默认 Logger 按 format 输出 Panic 日志后触发 panic。
func Panicf(ctx context.Context, format string, v ...interface{}) {
	Default().Panicf(ctx, format, v...)
}

// PrintStack 使用默认 Logger 打印调用栈。
func PrintStack(ctx context.Context, skip ...int) { Default().PrintStack(ctx, skip...) }

// ErrorMsg 兼容旧用法：将 message 与 error 合并为一条 Error 日志。
func ErrorMsg(ctx context.Context, message string, err error) {
	if err == nil {
		Errorf(ctx, "%s", message)
		return
	}
	Errorf(ctx, "%s ,err: %s", message, err.Error())
}

// ---- HTTP 辅助 ----

// RequestLogging 记录 GoFrame 请求处理摘要（方法、URL、头、状态、错误）。
func RequestLogging(o *ghttp.Request, err error) {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("\r\n📌 ---------------> [请求处理] Start %s  %s \r\n", o.Method, o.Header.Get("Content-Type")))
	sb.WriteString(fmt.Sprintf("    Origin  : %s \r\n", o.URL.String()))
	var header strings.Builder
	for k, v := range o.Header {
		if !gstr.Equal(k, "Content-Type") {
			header.WriteString(k + "=" + gstr.Join(v, ",") + "  ")
		}
	}
	if header.Len() > 0 {
		sb.WriteString(fmt.Sprintf("    Headers : %s \r\n", header.String()))
	}
	sb.WriteString(fmt.Sprintf("    Response: %d \r\n", o.Response.Status))
	if err != nil {
		sb.WriteString(fmt.Sprintf("    Error   : %s \r\n", err.Error()))
	}
	sb.WriteString(fmt.Sprintf("📌 ---------------> [请求处理] END %s \r\n", gstr.ToUpper(o.Proto)))
	Errorf(o.Context(), "%s", sb.String())
}

// formatHTTP 拼装 HTTP 调用调试文本。
func formatHTTP(method, url, contentType string, headers, response, param interface{}, err error) string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("\n------> %s  %s\n", method, url))
	if headers != nil {
		b.WriteString(fmt.Sprintf("Headers: %+v \n", headers))
	}
	if contentType != "" {
		b.WriteString(fmt.Sprintf("ContentType: %s \n", contentType))
	}
	if param != nil {
		b.WriteString(fmt.Sprintf("Params: %+v \n", param))
	}
	b.WriteString(fmt.Sprintf("Response: %s \n", response))
	if err != nil {
		b.WriteString(fmt.Sprintf("Error: %+v \n", err))
	}
	b.WriteString("------> END HTTP\n")
	return b.String()
}

// Call 以 Debug 级别输出一次 HTTP 调用信息（不含 error）。
func Call(ctx context.Context, method string, url string, contentType string, headers interface{}, response interface{}, param interface{}) {
	Debug(ctx, formatHTTP(method, url, contentType, headers, response, param, nil))
}

// DebugHTTP 在 isDebug 为 true 时输出 HTTP 调试信息（可含 error）。
// 命名避开包级 Debug，避免与级别方法冲突。
func DebugHTTP(ctx context.Context, isDebug bool, method string, url string, contentType string, headers interface{}, response interface{}, param interface{}, err error) {
	if !isDebug {
		return
	}
	Debug(ctx, formatHTTP(method, url, contentType, headers, response, param, err))
}
