package ctx

import (
	"context"

	"github.com/hosgf/element/types"
)

func Context(ctx context.Context) context.Context {
	if ctx != nil {
		return ctx
	}
	return context.Background()
}

// GetTenantId 从context中获取TenantId
func GetTenantId(ctx context.Context) string {
	return GetValue(ctx, types.TenantIdKey)
}

// GetTraceId 从context中获取TraceId
func GetTraceId(ctx context.Context) string {
	return GetValue(ctx, types.TraceIdKey)
}

// GetReqId 从context中获取ReqId
func GetReqId(ctx context.Context) string {
	return GetValue(ctx, types.RequestIdKey)
}

// GetUserId 从context中获取UserId
func GetUserId(ctx context.Context) string {
	return GetValue(ctx, types.UserIdKey)
}

// WithTenantId 将TenantId添加到context中
func WithTenantId(ctx context.Context, data string) context.Context {
	return WithValue(ctx, types.TenantIdKey, data)
}

// WithTraceId 将TraceId添加到context中
func WithTraceId(ctx context.Context, data string) context.Context {
	return WithValue(ctx, types.TraceIdKey, data)
}

// WithReqId 将ReqId添加到context中
func WithReqId(ctx context.Context, data string) context.Context {
	return WithValue(ctx, types.RequestIdKey, data)
}

// WithUserId 将UserId添加到context中
func WithUserId(ctx context.Context, userID string) context.Context {
	return WithValue(ctx, types.UserIdKey, userID)
}

func WithValue(ctx context.Context, key, value string) context.Context {
	ctx = Context(ctx)
	if value == "" {
		return ctx
	}
	if val, ok := ctx.Value(key).(string); ok && val != "" {
		return ctx
	}
	return context.WithValue(ctx, key, value)
}

func GetValue(ctx context.Context, key string) string {
	if ctx == nil {
		return ""
	}
	val, _ := ctx.Value(key).(string)
	return val
}
