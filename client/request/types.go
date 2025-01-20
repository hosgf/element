package request

import (
	"strings"

	"github.com/gin-gonic/gin"
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
	HeaderReqAppCode Header = "X-Req-App-Code"
	HeaderReqAppName Header = "X-Req-App-Name"
	HeaderReqClient  Header = "X-Req-Client"
	HeaderTraceId    Header = "X-Req-Id"
	HeaderUserAgent  Header = "X-User-Agent"
	HeaderReqToken   Header = "Authorization"
)

func GetHeaders() []Header {
	return []Header{HeaderReqAppCode, HeaderReqAppName, HeaderReqClient, HeaderTraceId, HeaderUserAgent, HeaderReqToken}
}

func (h Header) String() string {
	return string(h)
}

func GetAppCode(context *gin.Context) string {
	return GetHeader(context, HeaderReqAppCode)
}

func GetToken(context *gin.Context) string {
	return GetHeader(context, HeaderReqToken)
}

func GetAppName(context *gin.Context) string {
	return GetHeader(context, HeaderReqAppName)
}

func GetReqClient(context *gin.Context) string {
	return GetHeader(context, HeaderReqClient)
}

func GetTraceId(context *gin.Context) string {
	return GetHeader(context, HeaderTraceId)
}

func GetUserAgent(context *gin.Context) string {
	return GetHeader(context, HeaderUserAgent)
}

func GetHeader(context *gin.Context, key Header) string {
	return context.GetHeader(key.String())
}
