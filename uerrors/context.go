package uerrors

import (
	"context"

	"github.com/hosgf/element/client/request"
)

// GetRequestID 从context中获取RequestID
func GetRequestID(ctx context.Context) string {
	if ctx == nil {
		return ""
	}

	// 尝试从context中获取request_id
	if requestID, ok := ctx.Value("request_id").(string); ok && requestID != "" {
		return requestID
	}

	// 尝试从context中获取X-Request-ID
	if requestID, ok := ctx.Value(request.HeaderTraceId).(string); ok && requestID != "" {
		return requestID
	}

	return ""
}

// WithRequestID 将RequestID添加到context中
func WithRequestID(ctx context.Context, requestID string) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	return context.WithValue(ctx, "request_id", requestID)
}

// WithUserID 将UserID添加到context中
func WithUserID(ctx context.Context, userID string) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	return context.WithValue(ctx, "user_id", userID)
}

// GetUserID 从context中获取UserID
func GetUserID(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	if userID, ok := ctx.Value("user_id").(string); ok {
		return userID
	}
	return ""
}

// WithError 将错误信息添加到context中（用于调试）
func WithError(ctx context.Context, err error) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	return context.WithValue(ctx, "error", err)
}

// GetError 从context中获取错误信息
func GetError(ctx context.Context) error {
	if ctx == nil {
		return nil
	}
	if err, ok := ctx.Value("error").(error); ok {
		return err
	}
	return nil
}
