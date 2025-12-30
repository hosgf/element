package request

import (
	"context"
	"strings"

	"github.com/gogf/gf/v2/util/gconv"
)

const (
	CookieKey string = "_cookies"
)

// Protocol 请求协议类型
type Protocol string

const (
	HTTPS Protocol = "HTTPS"
	HTTP  Protocol = "HTTP"
	WS    Protocol = "WS"
)

func (t Protocol) String() string {
	return strings.ToUpper(string(t))
}

type Header string

const (
	HeaderReqAppCode   Header = "X-Req-App-Code"
	HeaderReqAppName   Header = "X-Req-App-Name"
	HeaderReqClient    Header = "X-Req-Client"
	HeaderSignature    Header = "X-Req-Secret"
	HeaderTraceId      Header = "X-Req-Id"
	HeaderUserAgent    Header = "X-User-Agent"
	HeaderTimestamp    Header = "timestamp"
	HeaderResponseTime Header = "X-Response-Time"
	HeaderReqToken     Header = "Authorization"
	HeaderSameToken    Header = "SA-SAME-TOKEN"
)

func GetHeaders() []Header {
	return []Header{HeaderReqAppCode, HeaderReqAppName, HeaderReqClient, HeaderTraceId, HeaderUserAgent, HeaderReqToken}
}

func (h Header) String() string {
	return string(h)
}

func (h Header) ToLowerString() string {
	return strings.ToLower(h.String())
}

func (h Header) Get(ctx context.Context) string {
	value := ctx.Value(h.String())
	if value == nil {
		return ""
	}
	return value.(string)
}

func SetSameToken(ctx context.Context, value interface{}) context.Context {
	if value == nil {
		return ctx
	}
	return context.WithValue(ctx, HeaderSameToken, gconv.String(value))
}

func GetHeaderList(ctx context.Context, keys ...string) map[string]string {
	headers := make(map[string]string)
	if len(keys) == 0 {
		return headers
	}
	for _, k := range keys {
		if value := Header(k).Get(ctx); len(value) > 0 {
			headers[k] = value
		}
	}
	return headers
}

func GetHeader(ctx context.Context) map[string]string {
	headers := make(map[string]string)
	for _, header := range GetHeaders() {
		if value := header.Get(ctx); len(value) > 0 {
			headers[header.String()] = value
		}
	}
	return headers
}

func SetHeader(ctx context.Context, key string, value interface{}) context.Context {
	if len(key) == 0 || value == nil {
		return ctx
	}
	return context.WithValue(ctx, key, gconv.String(value))
}

func SetHeaders(ctx context.Context, headers map[string]interface{}) context.Context {
	if headers == nil || len(headers) == 0 {
		return ctx
	}
	for k, v := range headers {
		if len(k) < 1 {
			continue
		}
		if v == nil {
			continue
		}
		ctx = context.WithValue(ctx, k, gconv.String(v))
	}
	return ctx
}

func GetCookies(ctx context.Context, cookie string) map[string]string {
	value := ctx.Value(cookie)
	if value == nil {
		return nil
	}
	return gconv.MapStrStr(value)
}

func SetCookies(ctx context.Context, cookie string, value map[string]string) context.Context {
	if value == nil || len(value) == 0 {
		return ctx
	}
	return context.WithValue(ctx, cookie, value)
}

func GetDefaultCookies(ctx context.Context) map[string]string {
	return GetCookies(ctx, CookieKey)
}

func SetDefaultCookies(ctx context.Context, value map[string]string) context.Context {
	return SetCookies(ctx, CookieKey, value)
}
