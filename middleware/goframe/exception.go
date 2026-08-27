package goframe

import (
	"context"
	"time"

	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/hosgf/element/client/request"
	"github.com/hosgf/element/ctx"
	"github.com/hosgf/element/types"

	"github.com/hosgf/element/logger"
	"github.com/hosgf/element/model/result"
	"github.com/hosgf/element/uerrors"
)

type recoverHandler struct {
	notify func(*uerrors.BizError)
}

var recoverDefault *recoverHandler

func initRecover() {
	recoverDefault = &recoverHandler{
		notify: func(err *uerrors.BizError) {
			logger.Errorf(context.Background(), "Global error notification: %s", err.Error())
		},
	}
}

func getRecover() *recoverHandler {
	if recoverDefault == nil {
		initRecover()
	}
	return recoverDefault
}

// SetNotify 设置全局错误通知回调。
func SetNotify(fn func(*uerrors.BizError)) {
	getRecover().notify = fn
}

// ensureIDs 补齐 trace / request，写入 ctx 与响应头（ApplyHTTP 幂等，Header 已设 trace 时仅补 request）。
func ensureIDs(r *ghttp.Request) {
	c := request.ApplyHTTP(r.Context(), GetHeader(r, request.HeaderTraceId), GetHeader(r, request.HeaderReqId))
	r.SetCtx(c)
	bindRespID(r, types.TraceIdKey, request.HeaderTraceId, ctx.GetTraceId(c))
	bindRespID(r, types.RequestIdKey, request.HeaderReqId, ctx.GetReqId(c))
}

func bindRespID(r *ghttp.Request, ctxKey string, header request.Header, id string) {
	if id == "" {
		return
	}
	r.SetCtxVar(ctxKey, id)
	r.SetCtxVar(header.String(), id)
	r.Response.Header().Set(header.String(), id)
}

// Recover 补齐请求标识、记录耗时、捕获 panic 并统一错误响应。
func Recover(r *ghttp.Request) {
	start := time.Now()
	ensureIDs(r)
	defer func() {
		r.Response.Header().Set(request.HeaderResponseTime.String(), time.Since(start).String())
		if v := recover(); v != nil {
			getRecover().handlePanic(r.Context(), r, v)
		}
	}()
	r.Middleware.Next()
}

// UseRecover 注册 Recover 中间件。
func UseRecover(server *ghttp.Server) *ghttp.Server {
	server.Use(Recover)
	return server
}

func (h *recoverHandler) handlePanic(ctx context.Context, r *ghttp.Request, v interface{}) {
	var bizErr *uerrors.BizError
	if err, ok := v.(error); ok {
		if be, ok := uerrors.IsBizError(err); ok {
			bizErr = be
		} else {
			bizErr = uerrors.WrapError(err, uerrors.ErrorTypeSystem, uerrors.ErrorLevelCritical, result.SC_FAILURE, "系统内部错误")
		}
	} else {
		bizErr = uerrors.NewBizError(
			uerrors.ErrorTypeSystem,
			uerrors.ErrorLevelCritical,
			result.SC_FAILURE,
			"系统内部错误",
			"panic",
		)
	}
	h.respond(ctx, r, bizErr)
}

func (h *recoverHandler) respond(reqCtx context.Context, r *ghttp.Request, bizErr *uerrors.BizError) {
	bizErr.RequestID = ctx.GetReqId(reqCtx)
	h.logErr(reqCtx, bizErr)
	if h.notify != nil {
		h.notify(bizErr)
	}
	h.writeErr(r, bizErr)
}

func (h *recoverHandler) logErr(ctx context.Context, err *uerrors.BizError) {
	msg := "[" + err.LevelString() + "] " + err.TypeString() + " - " + err.Message
	if err.Details != "" {
		msg += " | Details: " + err.Details
	}
	switch err.Level {
	case uerrors.ErrorLevelInfo:
		logger.Infof(ctx, "%s", msg)
	case uerrors.ErrorLevelWarning:
		logger.Warningf(ctx, "%s", msg)
	default:
		logger.Errorf(ctx, "%s", msg)
	}
}

func (h *recoverHandler) writeErr(r *ghttp.Request, err *uerrors.BizError) {
	response := result.NewResponse()
	response.Code = err.Code
	response.Message = err.Message
	result.Writer(r, response)
}
