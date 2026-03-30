package ctx

import (
	"context"

	"github.com/hosgf/element/types"
)

// GetReqId 从context中获取ReqId
func GetReqId(ctx context.Context) string {
	return getValue(ctx, types.RequestIdKey)
}

// GetTraceId 从context中获取TraceId
func GetTraceId(ctx context.Context) string {
	return getValue(ctx, types.TraceIdKey)
}

// GetTenantId 从context中获取TenantId
func GetTenantId(ctx context.Context) string {
	return getValue(ctx, types.TenantIdKey)
}

// GetUserId 从context中获取UserId
func GetUserId(ctx context.Context) string {
	return getValue(ctx, types.UserIdKey)
}

// WithReqId 将ReqId添加到context中
func WithReqId(ctx context.Context, data string) context.Context {
	return withValue(ctx, types.RequestIdKey, data)
}

// WithTraceId 将TraceId添加到context中
func WithTraceId(ctx context.Context, data string) context.Context {
	return withValue(ctx, types.TraceIdKey, data)
}

// WithTenantId 将TenantId添加到context中
func WithTenantId(ctx context.Context, data string) context.Context {
	return withValue(ctx, types.TenantIdKey, data)
}

// WithUserId 将UserId添加到context中
func WithUserId(ctx context.Context, userID string) context.Context {
	return withValue(ctx, types.UserIdKey, userID)
}

func getValue(ctx context.Context, key string) string {
	if ctx == nil {
		return ""
	}

	if val, ok := ctx.Value(key).(string); ok && val != "" {
		return val
	}
	return ""
}

func withValue(ctx context.Context, key, value string) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	if val, ok := ctx.Value(key).(string); ok && val != "" {
		return ctx
	}
	return context.WithValue(ctx, key, value)
}
