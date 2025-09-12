package goframe

import (
	"context"
	"time"

	"github.com/gogf/gf/v2/net/ghttp"

	"github.com/hosgf/element/config"
	"github.com/hosgf/element/exception"
	"github.com/hosgf/element/logger"
	"github.com/hosgf/element/model/result"
)

// exceptionHandler GoFrame 的全局异常处理器
type exceptionHandler struct {
	isProduction     bool
	enableStackTrace bool
	errorNotifier    func(*exception.BizError)
}

var handler *exceptionHandler

func initHandler() {
	isProduction := !config.Debug
	handler = &exceptionHandler{
		isProduction:     isProduction,
		enableStackTrace: !isProduction,
	}
	handler.SetErrorNotifier(func(err *exception.BizError) {
		logger.Errorf(context.Background(), "Global error notification: %s", err.Error())
	})
}

func getHandler() *exceptionHandler {
	if handler == nil {
		initHandler()
	}
	return handler
}

func (h *exceptionHandler) SetErrorNotifier(notifier func(*exception.BizError)) {
	h.errorNotifier = notifier
}

// ExceptionHandler GoFrame 异常中间件
func ExceptionHandler(r *ghttp.Request) {
	start := time.Now()
	if r.GetCtxVar("request_id").String() == "" {
		requestID := exception.GenerateRequestID()
		r.SetCtxVar("request_id", requestID)
		r.Response.Header().Set("X-Request-ID", requestID)
	}
	defer func() {
		r.Response.Header().Set("X-Response-Time", time.Since(start).String())
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
	requestID := r.GetCtxVar("request_id").String()
	if requestID == "" {
		requestID = "unknown"
	}
	var bizErr *exception.BizError
	if err, ok := panicValue.(error); ok {
		if be, isBiz := exception.IsBizError(err); isBiz {
			bizErr = be
		} else {
			bizErr = exception.WrapError(err, exception.ErrorTypeSystem, exception.ErrorLevelCritical, result.SC_FAILURE, "系统内部错误")
		}
	} else {
		bizErr = exception.NewBizError(
			exception.ErrorTypeSystem,
			exception.ErrorLevelCritical,
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
	requestID := r.GetCtxVar("request_id").String()
	if requestID == "" {
		requestID = "unknown"
	}
	var bizErr *exception.BizError
	if be, isBiz := exception.IsBizError(err); isBiz {
		bizErr = be
	} else {
		bizErr = exception.WrapError(err, exception.ErrorTypeSystem, exception.ErrorLevelError, result.SC_FAILURE, "系统错误")
	}
	bizErr.RequestID = requestID
	h.logError(ctx, bizErr)
	if h.errorNotifier != nil {
		h.errorNotifier(bizErr)
	}
	h.writeErrorResponse(r, bizErr)
}

func (h *exceptionHandler) logError(ctx context.Context, err *exception.BizError) {
	logMsg := "[" + err.LevelString() + "] " + err.TypeString() + " - " + err.Message
	if err.Details != "" {
		logMsg += " | Details: " + err.Details
	}
	if err.RequestID != "" {
		logMsg += " | RequestID: " + err.RequestID
	}
	switch err.Level {
	case exception.ErrorLevelInfo:
		logger.Log().Infof(ctx, logMsg)
	case exception.ErrorLevelWarning:
		logger.Warningf(ctx, logMsg)
	case exception.ErrorLevelError:
		logger.Errorf(ctx, logMsg)
	case exception.ErrorLevelCritical:
		logger.Errorf(ctx, logMsg)
	}
}

func (h *exceptionHandler) writeErrorResponse(r *ghttp.Request, err *exception.BizError) {
	// 仅返回顶层 code 与 message
	response := result.NewResponse()
	response.Code = err.Code
	response.Message = err.Message
	result.Writer(r, response)
}
