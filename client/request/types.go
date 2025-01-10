package request

import "github.com/gin-gonic/gin"

type Header string

const (
	HeaderReqAppCode Header = "X-Req-App-Code"
	HeaderReqAppName Header = "X-Req-App-Name"
	HeaderReqClient  Header = "X-Req-Client"
	HeaderTraceId    Header = "X-Req-Id"
	HeaderUserAgent  Header = "X-User-Agent"
)

func (h Header) String() string {
	return string(h)
}

func GetAppCode(context *gin.Context) string {
	return GetHeader(context, HeaderReqAppCode)
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
