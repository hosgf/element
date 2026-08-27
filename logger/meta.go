package logger

import (
	"context"
	"strings"

	"github.com/gogf/gf/v2/os/glog"
)

// 与 types / ctx 写入的 key 一致（logger 不可 import types：types→util→logger 环依赖）。
const (
	keyTrace = "trace_id"
	keyReq   = "request_id"
)

// MetaHandler 将 ctx 的 trace_id / request_id 写入 glog 头：{trace_id} <request_id>。
// 无 request_id 时不输出 <>；不在此生成 ID。
func MetaHandler(ctx context.Context, in *glog.HandlerInput) {
	if tid := ctxStr(ctx, keyTrace); tid != "" {
		in.TraceId = tid
	}
	if rid := ctxStr(ctx, keyReq); rid != "" {
		tag := "<" + rid + ">"
		if in.Prefix != "" {
			in.Prefix = tag + " " + in.Prefix
		} else {
			in.Prefix = tag
		}
	}
	in.Next(ctx)
}

// bindMeta 挂 MetaHandler，并清空 CtxKeys（避免 {a, b} 旧形态）。
func bindMeta(l *glog.Logger) *glog.Logger {
	if l == nil {
		return nil
	}
	l.SetCtxKeys()
	l.SetHandlers(MetaHandler)
	return l
}

func ctxStr(ctx context.Context, key string) string {
	if ctx == nil {
		return ""
	}
	v, _ := ctx.Value(key).(string)
	return strings.TrimSpace(v)
}
