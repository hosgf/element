package ugin

import (
	"context"
	"time"

	gingonic "github.com/gin-gonic/gin"

	"github.com/hosgf/element/config"
	"github.com/hosgf/element/logger"
	"github.com/hosgf/element/model/result"
	"github.com/hosgf/element/uerrors"
)

// exceptionHandler Gin 的全局异常处理器
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

// ExceptionHandler 返回 Gin 中间件
func ExceptionHandler() gingonic.HandlerFunc {
	h := getHandler()
	return func(c *gingonic.Context) {
		start := time.Now()
		if c.GetString("request_id") == "" {
			requestID := uerrors.GenerateRequestID()
			c.Set("request_id", requestID)
			c.Writer.Header().Set("X-Request-ID", requestID)
		}
		defer func() {
			c.Writer.Header().Set("X-Response-Time", time.Since(start).String())
			if panicValue := recover(); panicValue != nil {
				h.HandlePanic(c.Request.Context(), c, panicValue)
				c.Abort()
			}
		}()
		c.Next()
	}
}

func (h *exceptionHandler) HandlePanic(ctx context.Context, c *gingonic.Context, panicValue interface{}) {
	requestID := c.GetString("request_id")
	if requestID == "" {
		requestID = c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = "unknown"
		}
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
			"panic",
		)
	}

	bizErr.RequestID = requestID

	h.logError(ctx, bizErr)
	if h.errorNotifier != nil {
		h.errorNotifier(bizErr)
	}
	h.writeErrorResponse(c, bizErr)
}

func (h *exceptionHandler) HandleError(ctx context.Context, c *gingonic.Context, err error) {
	requestID := c.GetString("request_id")
	if requestID == "" {
		requestID = c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = "unknown"
		}
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
	h.writeErrorResponse(c, bizErr)
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

func (h *exceptionHandler) writeErrorResponse(c *gingonic.Context, err *uerrors.BizError) {
	// 仅返回顶层 code 与 message
	response := result.NewResponse()
	response.Code = err.Code
	response.Message = err.Message
	c.Status(200)
	c.JSON(200, response)
}
