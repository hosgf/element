package ctx

import (
	"context"
	"strings"

	"github.com/hosgf/element/client/request"
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

// GetClientIP 从context中获取网关传递的客户端真实IP。
func GetClientIP(ctx context.Context) string {
	return getValue(ctx, types.ClientIPKey, request.HeaderRealIP.String())
}

// GetDeviceCode 从context中获取设备编码。
func GetDeviceCode(ctx context.Context) string {
	return getValue(ctx, types.DeviceCodeKey, request.HeaderDeviceCode.String())
}

// GetDeviceType 从context中获取设备类型。
func GetDeviceType(ctx context.Context) string {
	return getValue(ctx, types.DeviceTypeKey, request.HeaderDeviceType.String())
}

// GetUserAgent 从context中获取客户端User-Agent。
func GetUserAgent(ctx context.Context) string {
	return getValue(ctx, types.UserAgentKey, request.HeaderUserAgent.String())
}

// GetReqClient 从context中获取请求客户端标识。
func GetReqClient(ctx context.Context) string {
	return getValue(ctx, types.ReqClientKey, request.HeaderReqClient.String())
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

// WithClientIP 将客户端真实IP添加到context中。
func WithClientIP(ctx context.Context, clientIP string) context.Context {
	return WithValue(ctx, types.ClientIPKey, strings.TrimSpace(clientIP))
}

// WithDeviceCode 将设备编码添加到context中。
func WithDeviceCode(ctx context.Context, deviceCode string) context.Context {
	return WithValue(ctx, types.DeviceCodeKey, strings.TrimSpace(deviceCode))
}

// WithDeviceType 将设备类型添加到context中。
func WithDeviceType(ctx context.Context, deviceType string) context.Context {
	return WithValue(ctx, types.DeviceTypeKey, strings.TrimSpace(deviceType))
}

// WithUserAgent 将客户端User-Agent添加到context中。
func WithUserAgent(ctx context.Context, userAgent string) context.Context {
	return WithValue(ctx, types.UserAgentKey, strings.TrimSpace(userAgent))
}

// WithReqClient 将请求客户端标识添加到context中。
func WithReqClient(ctx context.Context, reqClient string) context.Context {
	return WithValue(ctx, types.ReqClientKey, strings.TrimSpace(reqClient))
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

func getValue(ctx context.Context, keys ...string) string {
	if ctx == nil {
		return ""
	}
	for _, key := range keys {
		if val := strings.TrimSpace(GetValue(ctx, key)); val != "" {
			return val
		}
	}
	return ""
}
