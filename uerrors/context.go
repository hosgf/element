package uerrors

import (
	"context"
	"strings"

	"github.com/hosgf/element/client/request"
	"github.com/hosgf/element/ctx"
	"github.com/hosgf/element/types"
)

// GetRequestID 从 context 获取 request_id（语义 key 优先，其次 X-Req-Id）。
func GetRequestID(c context.Context) string {
	if rid := ctx.GetReqId(c); rid != "" {
		return rid
	}
	return strings.TrimSpace(request.HeaderReqId.Get(c))
}

// WithRequestID 将 RequestID 写入 context。
func WithRequestID(c context.Context, requestID string) context.Context {
	if c == nil {
		c = context.Background()
	}
	return context.WithValue(c, types.RequestIdKey, strings.TrimSpace(requestID))
}

// WithUserID 将 UserID 写入 context。
func WithUserID(c context.Context, userID string) context.Context {
	if c == nil {
		c = context.Background()
	}
	return context.WithValue(c, types.UserIdKey, userID)
}

// GetUserID 从 context 获取 UserID。
func GetUserID(c context.Context) string {
	if c == nil {
		return ""
	}
	if userID, ok := c.Value(types.UserIdKey).(string); ok {
		return userID
	}
	return ""
}

// WithError 将错误信息写入 context（调试用）。
func WithError(c context.Context, err error) context.Context {
	if c == nil {
		c = context.Background()
	}
	return context.WithValue(c, "error", err)
}

// GetError 从 context 获取错误信息。
func GetError(c context.Context) error {
	if c == nil {
		return nil
	}
	if err, ok := c.Value("error").(error); ok {
		return err
	}
	return nil
}
