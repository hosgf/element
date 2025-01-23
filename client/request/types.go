package request

import (
	"strings"
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
