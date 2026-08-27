package middleware

import (
	"github.com/hosgf/element/client/request"
	"github.com/hosgf/element/types"
)

// HeaderBind 语义 ctx key 与 HTTP header 的绑定。
type HeaderBind struct {
	Key    string
	Header request.Header
}

// ContextHeaders 需写入语义 key 的请求头（trace 由 bindTrace 单独处理）。
var ContextHeaders = []HeaderBind{
	{types.RequestIdKey, request.HeaderReqId},
	{types.TenantIdKey, request.HeaderTenantId},
	{types.UserIdKey, request.HeaderUserId},
	{types.ClientIPKey, request.HeaderRealIP},
	{types.DeviceCodeKey, request.HeaderDeviceCode},
	{types.DeviceTypeKey, request.HeaderDeviceType},
	{types.UserAgentKey, request.HeaderUserAgent},
	{types.ReqClientKey, request.HeaderReqClient},
}
