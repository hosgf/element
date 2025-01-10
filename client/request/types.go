package request

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
