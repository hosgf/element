package result

import (
	"compress/gzip"
	"encoding/json"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/net/ghttp"
)

type Response struct {
	Code    int         `json:"code,omitempty"`    // 结果码
	Message string      `json:"message,omitempty"` // 消息
	Data    interface{} `json:"data,omitempty"`    // 数据集
}

var response = new(Response)

func NewResponse() *Response {
	return &Response{}
}

func Build(code int, message string, data interface{}) *Response {
	return &Response{
		Code:    code,
		Message: message,
		Data:    data,
	}
}

func Download(r *ghttp.Request, path string, err error) {
	if err != nil {
		r.Response.WriteStatus(404)
		r.Response.WriteExit(err.Error())
	} else {
		r.Response.ServeFileDownload(path)
	}
}

func Success(r *ghttp.Request, data interface{}) {
	response.Success(r, data)
}

func Fail(r *ghttp.Request, err error) {
	response.Fail(r, SC_FAILURE, err)
}

func (res *Response) Fail(r *ghttp.Request, code int, err error) {
	if nil == err {
		return
	}
	res.Build(r, code, err.Error(), nil)
}

func (res *Response) Success(r *ghttp.Request, data interface{}) {
	res.Result(r, SUCCESS, data)
}

func (res *Response) Result(r *ghttp.Request, resultCode gcode.Code, data interface{}) {
	res.Build(r, resultCode.Code(), resultCode.Message(), data)
}

func (res *Response) Build(r *ghttp.Request, code int, message string, data interface{}) {
	res.gzip(r, Build(code, message, data))
}

func (res *Response) gzip(r *ghttp.Request, data *Response) {
	r.Response.Header().Set("Content-Type", "application/json")
	r.Response.Header().Set("Content-Encoding", "gzip")
	gw := gzip.NewWriter(r.Response.Writer)
	defer gw.Close()
	_ = json.NewEncoder(gw).Encode(data)
	r.Exit()
}

func Writer(r *ghttp.Request, data interface{}) {
	if nil == data {
		return
	}
	r.Response.Header().Set("Content-Type", "application/json")
	r.Response.Header().Set("Content-Encoding", "gzip")
	gw := gzip.NewWriter(r.Response.Writer)
	defer gw.Close()
	_ = json.NewEncoder(gw).Encode(data)
	r.Exit()
}
