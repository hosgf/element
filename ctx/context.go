package ctx

import (
	"context"

	"github.com/hosgf/element/client/request"
	"github.com/hosgf/element/types"
)

// GetReqId 从context中获取ReqId
func GetReqId(ctx context.Context) string {
	if ctx == nil {
		return ""
	}

	// 尝试从context中获取request_id
	if requestID, ok := ctx.Value(types.RequestIdKey).(string); ok && requestID != "" {
		return requestID
	}

	// 尝试从context中获取X-Request-ID
	if requestID, ok := ctx.Value(request.HeaderTraceId).(string); ok && requestID != "" {
		return requestID
	}

	return ""
}

// GetUserId 从context中获取UserId
func GetUserId(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	if userID, ok := ctx.Value(types.UserIdKey).(string); ok {
		return userID
	}
	return ""
}

// WithReqId 将ReqId添加到context中
func WithReqId(ctx context.Context, requestID string) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	return context.WithValue(ctx, types.RequestIdKey, requestID)
}

// WithUserId 将UserId添加到context中
func WithUserId(ctx context.Context, userID string) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	return context.WithValue(ctx, types.UserIdKey, userID)
}
