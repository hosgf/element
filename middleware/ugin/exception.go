package ugin

import (
	"context"
	"time"

	gingonic "github.com/gin-gonic/gin"
	"github.com/hosgf/element/client/request"
	"github.com/hosgf/element/ctx"
	"github.com/hosgf/element/trace"
	"github.com/hosgf/element/types"

	"github.com/hosgf/element/logger"
	"github.com/hosgf/element/model/result"
	"github.com/hosgf/element/uerrors"
)

type exceptionHandler struct {
	notify func(*uerrors.BizError)
}

var handler *exceptionHandler

func initHandler() {
	handler = &exceptionHandler{
		notify: func(err *uerrors.BizError) {
			logger.Errorf(context.Background(), "Global error notification: %s", err.Error())
		},
	}
}

func getHandler() *exceptionHandler {
	if handler == nil {
		initHandler()
	}
	return handler
}

// SetNotify 设置全局错误通知回调。
func SetNotify(fn func(*uerrors.BizError)) {
	getHandler().notify = fn
}

// ExceptionHandler 返回 Gin 中间件。
func ExceptionHandler() gingonic.HandlerFunc {
	h := getHandler()
	return func(c *gingonic.Context) {
		start := time.Now()
		ensureIDs(c)
		defer func() {
			c.Writer.Header().Set(request.HeaderResponseTime.String(), time.Since(start).String())
			if v := recover(); v != nil {
				h.handlePanic(c.Request.Context(), c, v)
				c.Abort()
			}
		}()
		c.Next()
	}
}

// ensureIDs 补齐 request_id，写入 gin + ctx，回写响应头。
func ensureIDs(c *gingonic.Context) {
	reqCtx := trace.Sync(request.EnsureID(c.Request.Context(), GetHeader(c, request.HeaderReqId)))
	c.Request = c.Request.WithContext(reqCtx)
	bindRespID(c, types.RequestIdKey, request.HeaderReqId, ctx.GetReqId(reqCtx))
}

func bindRespID(c *gingonic.Context, key string, header request.Header, id string) {
	if id == "" {
		return
	}
	reqCtx := bindID(c.Request.Context(), c, key, header, id)
	c.Request = c.Request.WithContext(reqCtx)
	c.Writer.Header().Set(header.String(), id)
}

func (h *exceptionHandler) handlePanic(ctx context.Context, c *gingonic.Context, v interface{}) {
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
	h.respond(ctx, c, bizErr)
}

// HandleError 显式处理业务错误并写统一响应。
func HandleError(ctx context.Context, c *gingonic.Context, err error) {
	getHandler().handleError(ctx, c, err)
}

func (h *exceptionHandler) handleError(ctx context.Context, c *gingonic.Context, err error) {
	var bizErr *uerrors.BizError
	if be, ok := uerrors.IsBizError(err); ok {
		bizErr = be
	} else {
		bizErr = uerrors.WrapError(err, uerrors.ErrorTypeSystem, uerrors.ErrorLevelError, result.SC_FAILURE, "系统错误")
	}
	h.respond(ctx, c, bizErr)
}

func (h *exceptionHandler) respond(reqCtx context.Context, c *gingonic.Context, bizErr *uerrors.BizError) {
	bizErr.RequestID = ctx.GetReqId(reqCtx)
	h.logErr(reqCtx, bizErr)
	if h.notify != nil {
		h.notify(bizErr)
	}
	h.writeErr(c, bizErr)
}

func (h *exceptionHandler) logErr(ctx context.Context, err *uerrors.BizError) {
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

func (h *exceptionHandler) writeErr(c *gingonic.Context, err *uerrors.BizError) {
	response := result.NewResponse()
	response.Code = err.Code
	response.Message = err.Message
	c.Status(200)
	c.JSON(200, response)
}
