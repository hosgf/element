package goframe

import (
	"context"
	"time"

	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/hosgf/element/client/request"
	"github.com/hosgf/element/types"

	"github.com/hosgf/element/config"
	"github.com/hosgf/element/logger"
	"github.com/hosgf/element/model/result"
	"github.com/hosgf/element/uerrors"
)

// exceptionHandler GoFrame 的全局异常处理器
type exceptionHandler struct {
	isProduction     bool
	enableStackTrace bool
	errorNotifier    func(*uerrors.BizError)
}

var handler *exceptionHandler

func initHandler() {
	isProduction := !config.Debug
	handler = &exceptionHandler{
		isProduction:     isProduction,
		enableStackTrace: !isProduction,
	}
	handler.SetErrorNotifier(func(err *uerrors.BizError) {
		logger.Errorf(context.Background(), "Global error notification: %s", err.Error())
	})
}

func getHandler() *exceptionHandler {
	if handler == nil {
		initHandler()
	}
	return handler
}

func (h *exceptionHandler) SetErrorNotifier(notifier func(*uerrors.BizError)) {
	h.errorNotifier = notifier
}

// ExceptionHandler GoFrame 异常中间件
func ExceptionHandler(r *ghttp.Request) {
	start := time.Now()
	if r.GetCtxVar(types.RequestIdKey).String() == "" {
		requestID := request.GenerateRequestID()
		r.SetCtxVar(types.RequestIdKey, requestID)
		r.Response.Header().Set(request.HeaderTraceId.String(), requestID)
	}
	defer func() {
		r.Response.Header().Set(request.HeaderResponseTime.String(), time.Since(start).String())
		if panicValue := recover(); panicValue != nil {
			getHandler().HandlePanic(r.Context(), r, panicValue)
		}
	}()
	r.Middleware.Next()
}

// UseException 绑定异常中间件
func UseException(server *ghttp.Server) *ghttp.Server {
	server.Use(ExceptionHandler)
	return server
}

func (h *exceptionHandler) HandlePanic(ctx context.Context, r *ghttp.Request, panicValue interface{}) {
	requestID := r.GetCtxVar(types.RequestIdKey).String()
	if requestID == "" {
		requestID = "unknown"
	}
	var bizErr *uerrors.BizError
	if err, ok := panicValue.(error); ok {
		if be, isBiz := uerrors.IsBizError(err); isBiz {
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
	bizErr.RequestID = requestID
	h.logError(ctx, bizErr)
	if h.errorNotifier != nil {
		h.errorNotifier(bizErr)
	}
	h.writeErrorResponse(r, bizErr)
}

func (h *exceptionHandler) HandleError(ctx context.Context, r *ghttp.Request, err error) {
	requestID := r.GetCtxVar(types.RequestIdKey).String()
	if requestID == "" {
		requestID = "unknown"
	}
	var bizErr *uerrors.BizError
	if be, isBiz := uerrors.IsBizError(err); isBiz {
		bizErr = be
	} else {
		bizErr = uerrors.WrapError(err, uerrors.ErrorTypeSystem, uerrors.ErrorLevelError, result.SC_FAILURE, "系统错误")
	}
	bizErr.RequestID = requestID
	h.logError(ctx, bizErr)
	if h.errorNotifier != nil {
		h.errorNotifier(bizErr)
	}
	h.writeErrorResponse(r, bizErr)
}

func (h *exceptionHandler) logError(ctx context.Context, err *uerrors.BizError) {
	logMsg := "[" + err.LevelString() + "] " + err.TypeString() + " - " + err.Message
	if err.Details != "" {
		logMsg += " | Details: " + err.Details
	}
	if err.RequestID != "" {
		logMsg += " | RequestID: " + err.RequestID
	}
	switch err.Level {
	case uerrors.ErrorLevelInfo:
		logger.Log().Infof(ctx, logMsg)
	case uerrors.ErrorLevelWarning:
		logger.Warningf(ctx, logMsg)
	case uerrors.ErrorLevelError:
		logger.Errorf(ctx, logMsg)
	case uerrors.ErrorLevelCritical:
		logger.Errorf(ctx, logMsg)
	}
}

func (h *exceptionHandler) writeErrorResponse(r *ghttp.Request, err *uerrors.BizError) {
	// 仅返回顶层 code 与 message
	response := result.NewResponse()
	response.Code = err.Code
	response.Message = err.Message
	result.Writer(r, response)
}
