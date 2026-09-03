package goframe

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/hosgf/element/client/request"
	"github.com/hosgf/element/logger"
	"github.com/hosgf/element/util"
)

const (
	defaultMaxParam = 4096
	defaultMaxResp  = 2000
	defaultMaxErr   = 12000

	// AccessFailureKey / AccessErrorKey 业务失败写入 ctx，AccessLog 默认读取。
	AccessFailureKey = "element_http_failure"
	AccessErrorKey   = "element_http_error"
)

var (
	multiSpaceRegex = regexp.MustCompile(`\s+`)
)

var defaultLogHeaders = []request.Header{
	request.HeaderTraceparent,
	request.HeaderReqId,
	request.HeaderUserAgent,
	request.HeaderContentType,
	request.HeaderReqAppName,
	request.HeaderReqAppCode,
	request.HeaderReqClient,
	request.HeaderTimestamp,
	request.HeaderSignature,
	request.HeaderTraceId,
}

var defaultRedact = []string{
	"password", "privatekey", "private_key", "secret", "token",
	"apikey", "api_key", "auth_salt", "authorization",
	"cookie", "session", "access_token", "refresh_token",
}

// AccessLogOptions 访问日志参数。不读配置文件，由调用方构造。
type AccessLogOptions struct {
	SkipPaths  []string
	MaxParam   int
	MaxResp    int
	Headers    []request.Header
	Redact     []string
	FailureKey string
	ErrorKey   string
	Info       func(ctx context.Context, msg string)
	Warn       func(ctx context.Context, msg string)
}

type accessLog struct {
	skipPaths   []string
	maxParam    int
	maxResp     int
	headers     []request.Header
	redact      []string
	redactSet   map[string]struct{}
	failureKey  string
	errorKey    string
	info        func(ctx context.Context, msg string)
	warn        func(ctx context.Context, msg string)
	redactCache sync.Map
}

// AccessLog 记录请求参数与响应预览（明文）。须注册在 MiddlewareGzip 内侧。
// 无参即可用：s.Use(AccessLog())
func AccessLog(o ...AccessLogOptions) ghttp.HandlerFunc {
	var opts AccessLogOptions
	if len(o) > 0 {
		opts = o[0]
	}
	return newAccessLog(opts).handle
}

func newAccessLog(o AccessLogOptions) *accessLog {
	if o.MaxParam <= 0 {
		o.MaxParam = defaultMaxParam
	}
	if o.MaxResp <= 0 {
		o.MaxResp = defaultMaxResp
	}
	if len(o.Headers) == 0 {
		o.Headers = defaultLogHeaders
	}
	if len(o.Redact) == 0 {
		o.Redact = defaultRedact
	}
	if o.Info == nil {
		o.Info = func(ctx context.Context, msg string) { logger.Default().Info(ctx, msg) }
	}
	if o.Warn == nil {
		o.Warn = func(ctx context.Context, msg string) { logger.Default().Warning(ctx, msg) }
	}
	if o.FailureKey == "" {
		o.FailureKey = AccessFailureKey
	}
	if o.ErrorKey == "" {
		o.ErrorKey = AccessErrorKey
	}
	skipPaths := o.SkipPaths
	if skipPaths == nil {
		skipPaths = []string{"/health", "/ping"}
	}
	redactSet := make(map[string]struct{}, len(o.Redact))
	for _, f := range o.Redact {
		redactSet[f] = struct{}{}
	}
	return &accessLog{
		skipPaths:  skipPaths,
		maxParam:   o.MaxParam,
		maxResp:    o.MaxResp,
		headers:    o.Headers,
		redact:     o.Redact,
		redactSet:  redactSet,
		failureKey: o.FailureKey,
		errorKey:   o.ErrorKey,
		info:       o.Info,
		warn:       o.Warn,
	}
}

func (a *accessLog) handle(r *ghttp.Request) {
	if a.skipped(r) {
		r.Middleware.Next()
		return
	}

	start := time.Now()
	var prefix strings.Builder
	n := contentLength(r)
	line := fmt.Sprintf("\r\n--> %s %s", r.Method, buildFullURL(r))
	if n > 0 {
		line += fmt.Sprintf(" (%d bytes)", n)
	}
	prefix.WriteString(line + "\n")

	var pairs []string
	for _, h := range a.headers {
		value := GetHeader(r, h)
		if value == "" {
			continue
		}
		value = a.redactHeader(h, value)
		pairs = append(pairs, fmt.Sprintf("%s=%s", h, value))
	}
	if len(pairs) > 0 {
		prefix.WriteString(fmt.Sprintf("  [Request] %s\n", strings.Join(pairs, " | ")))
	}
	if params := a.formatParams(r); params != "" {
		prefix.WriteString("  [Params] " + params + "\n")
	}

	r.Middleware.Next()

	status := r.Response.Status
	statusText := http.StatusText(status)
	if statusText == "" {
		statusText = "Unknown"
	}

	var out strings.Builder
	out.WriteString(prefix.String())

	body := decodeBody(r.Response.Buffer())
	length := int64(len(r.Response.Buffer()))
	if length == 0 {
		length = r.Response.BytesWritten()
	}
	if body != "" {
		out.WriteString(fmt.Sprintf("  [Response] %s\n", a.truncate(body, a.maxResp)))
	} else if length > 0 {
		out.WriteString(fmt.Sprintf("  [Response] <unavailable> (%d bytes)\n", length))
	}

	ms := time.Since(start).Milliseconds()
	out.WriteString(fmt.Sprintf("<-- END HTTP %d %s (%dms, %d bytes)\n", status, statusText, ms, length))

	failed := false
	if a.failureKey != "" {
		if raw := contextString(r, a.failureKey); raw != "" {
			failed = true
			a.writeFailure(&out, raw, contextError(r, a.errorKey))
		}
	}

	msg := out.String()
	if failed || status < 200 || status >= 300 {
		a.warn(r.Context(), msg)
	} else {
		a.info(r.Context(), msg)
	}
}

func (a *accessLog) skipped(r *ghttp.Request) bool {
	if r == nil || r.URL == nil {
		return false
	}
	return util.MatchPath(r.URL.Path, a.skipPaths)
}

func (a *accessLog) formatParams(r *ghttp.Request) string {
	contentType := strings.ToLower(GetHeader(r, request.HeaderContentType))
	if strings.Contains(contentType, "multipart/form-data") {
		if n := contentLength(r); n > 0 {
			return fmt.Sprintf("multipart (%d bytes)", n)
		}
		return "multipart"
	}
	body := strings.TrimSpace(r.GetBodyString())
	if body != "" {
		if strings.Contains(contentType, "application/octet-stream") {
			return fmt.Sprintf("binary body (%d bytes)", len(body))
		}
		return "body=" + a.truncate(body, a.maxParam)
	}
	if q := r.URL.RawQuery; q != "" {
		return "query=" + a.truncate(q, a.maxParam)
	}
	return ""
}

func (a *accessLog) truncate(raw string, maxLen int) string {
	raw = strings.TrimSpace(multiSpaceRegex.ReplaceAllString(raw, " "))
	raw = a.redactText(raw)
	if maxLen > 0 && len(raw) > maxLen {
		raw = raw[:maxLen] + "...(truncated)"
	}
	return raw
}

func (a *accessLog) writeFailure(out *strings.Builder, raw, detail string) {
	raw = strings.TrimSpace(raw)
	msg := raw
	var parsed struct {
		Message string `json:"message"`
	}
	if json.Unmarshal([]byte(raw), &parsed) == nil && parsed.Message != "" {
		msg = parsed.Message
	}
	out.WriteString(fmt.Sprintf("  [异常] message=%s\n", a.truncate(msg, 0)))
	if detail != "" {
		detail = a.redactText(detail)
		if len(detail) > defaultMaxErr {
			detail = detail[:defaultMaxErr] + "...(truncated)"
		}
		for _, line := range strings.Split(detail, "\n") {
			out.WriteString("    " + line + "\n")
		}
	}
}

func (a *accessLog) redactHeader(key request.Header, value string) string {
	if value == "" {
		return value
	}
	lower := key.ToLowerString()
	for field := range a.redactSet {
		if strings.Contains(lower, field) {
			return redactValue(value)
		}
	}
	return value
}

func (a *accessLog) redactText(text string) string {
	if text == "" {
		return text
	}
	lower := strings.ToLower(text)
	hit := false
	for field := range a.redactSet {
		if strings.Contains(lower, field) {
			hit = true
			break
		}
	}
	if !hit {
		return text
	}
	for _, field := range a.redact {
		if !strings.Contains(lower, field) {
			continue
		}
		re := a.redactPattern(field)
		text = re.ReplaceAllStringFunc(text, func(match string) string {
			parts := re.FindStringSubmatch(match)
			if len(parts) >= 4 {
				return parts[1] + redactValue(parts[2]) + parts[3]
			}
			return match
		})
		lower = strings.ToLower(text)
	}
	return text
}

func (a *accessLog) redactPattern(field string) *regexp.Regexp {
	if cached, ok := a.redactCache.Load(field); ok {
		return cached.(*regexp.Regexp)
	}
	pattern := fmt.Sprintf(`(?i)("%s"\s*:\s*")([^"]+)(")`, regexp.QuoteMeta(field))
	re := regexp.MustCompile(pattern)
	actual, _ := a.redactCache.LoadOrStore(field, re)
	return actual.(*regexp.Regexp)
}

func redactValue(value string) string {
	if len(value) <= 8 {
		return "***REDACTED***"
	}
	return value[:4] + "***REDACTED***" + value[len(value)-4:]
}

func decodeBody(raw []byte) string {
	if len(raw) == 0 {
		return ""
	}
	if len(raw) >= 2 && raw[0] == 0x1f && raw[1] == 0x8b {
		gr, err := gzip.NewReader(bytes.NewReader(raw))
		if err != nil {
			return fmt.Sprintf("<gzip %d bytes>", len(raw))
		}
		defer gr.Close()
		plain, err := io.ReadAll(io.LimitReader(gr, int64(defaultMaxResp)*4))
		if err != nil && err != io.EOF {
			return fmt.Sprintf("<gzip %d bytes>", len(raw))
		}
		return string(plain)
	}
	return string(raw)
}

func buildFullURL(r *ghttp.Request) string {
	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}
	host := r.Host
	if host == "" {
		host = r.Header.Get("Host")
	}
	if host == "" {
		host = "localhost"
	}
	return fmt.Sprintf("%s://%s%s", scheme, host, r.RequestURI)
}

func contentLength(r *ghttp.Request) int64 {
	if s := r.Header.Get("Content-Length"); s != "" {
		if n, err := strconv.ParseInt(s, 10, 64); err == nil {
			return n
		}
	}
	return 0
}

func contextString(r *ghttp.Request, key string) string {
	if key == "" {
		return ""
	}
	if v := r.GetCtxVar(key); v != nil && !v.IsNil() {
		return v.String()
	}
	return ""
}

func contextError(r *ghttp.Request, key string) string {
	if key == "" {
		return ""
	}
	if v := r.GetCtxVar(key); v != nil && !v.IsNil() {
		if err, ok := v.Val().(error); ok && err != nil {
			return fmt.Sprintf("%+v", err)
		}
	}
	return ""
}
