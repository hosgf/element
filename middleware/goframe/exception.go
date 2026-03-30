package goframe

import (
	"context"
	"time"

	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/hosgf/element/client/request"
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
	recoverDefault = &recoverHandler{}
	recoverDefault.SetNotify(func(err *uerrors.BizError) {
		logger.Errorf(context.Background(), "Global error notification: %s", err.Error())
	})
}

func getRecover() *recoverHandler {
	if recoverDefault == nil {
		initRecover()
	}
	return recoverDefault
}

func (h *recoverHandler) SetNotify(fn func(*uerrors.BizError)) {
	h.notify = fn
}

func ensureIDs(r *ghttp.Request) {
	bindCtxHeader(r, types.TraceIdKey, request.HeaderTraceId, request.GenerateRequestID)
	bindCtxHeader(r, types.RequestIdKey, request.HeaderReqId, request.GenerateRequestID)
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
		)
	}
	h.respond(ctx, r, bizErr)
}

func (h *recoverHandler) handleErr(ctx context.Context, r *ghttp.Request, err error) {
	var bizErr *uerrors.BizError
	if be, ok := uerrors.IsBizError(err); ok {
		bizErr = be
	} else {
		bizErr = uerrors.WrapError(err, uerrors.ErrorTypeSystem, uerrors.ErrorLevelError, result.SC_FAILURE, "系统错误")
	}
	h.respond(ctx, r, bizErr)
}

func (h *recoverHandler) respond(ctx context.Context, r *ghttp.Request, bizErr *uerrors.BizError) {
	requestID := r.GetCtxVar(types.RequestIdKey).String()
	if requestID == "" {
		requestID = "unknown"
	}
	bizErr.RequestID = requestID
	h.logErr(ctx, bizErr)
	if h.notify != nil {
		h.notify(bizErr)
	}
	h.writeErr(r, bizErr)
}

func (h *recoverHandler) logErr(ctx context.Context, err *uerrors.BizError) {
	logMsg := "[" + err.LevelString() + "] " + err.TypeString() + " - " + err.Message
	if err.Details != "" {
		logMsg += " | Details: " + err.Details
	}
	if err.RequestID != "" {
		logMsg += " | RequestID: " + err.RequestID
	}
	switch err.Level {
	case uerrors.ErrorLevelInfo:
		logger.Log().Infof(ctx, "%s", logMsg)
	case uerrors.ErrorLevelWarning:
		logger.Warningf(ctx, "%s", logMsg)
	case uerrors.ErrorLevelError, uerrors.ErrorLevelCritical:
		logger.Errorf(ctx, "%s", logMsg)
	}
}

func (h *recoverHandler) writeErr(r *ghttp.Request, err *uerrors.BizError) {
	response := result.NewResponse()
	response.Code = err.Code
	response.Message = err.Message
	result.Writer(r, response)
}

func bindCtxHeader(r *ghttp.Request, ctxKey string, header request.Header, defaultID func() string) {
	if r.GetCtxVar(ctxKey).String() != "" {
		return
	}
	headerVal := GetHeader(r, header)
	var id string
	if len(headerVal) > 0 {
		id = headerVal
	} else if defaultID != nil {
		id = defaultID()
	}
	if id == "" {
		return
	}
	r.SetCtxVar(ctxKey, id)
	r.Response.Header().Set(header.String(), id)
}
