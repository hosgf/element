package result

import "github.com/gogf/gf/v2/errors/gcode"

const (
	SC_OK                  = 200
	SC_BAD_REQUEST         = 400
	SC_UNAUTHORIZED        = 401
	SC_FORBIDDEN           = 403
	SC_NOT_FOUND           = 404
	SC_TIMEOUT             = 408
	SC_FAILURE             = 500
	SC_GATEWAY             = 502
	SC_SERVICE_UNAVAILABLE = 503
	SC_INTERNAL_ERROR      = 506
	SC_SERVICE_ERROR       = 5700
	SC_UNSUPPORTED_ERROR   = 5701
	SC_UPLOAD_ERROR        = 5702
)

var (
	SUCCESS           = gcode.New(SC_OK, "操作成功", "")
	NOT_FOUND         = gcode.New(SC_NOT_FOUND, "资源不存在", "")
	REQ_REJECT        = gcode.New(SC_FORBIDDEN, "请求被拒绝", "")
	UN_AUTHORIZED     = gcode.New(SC_UNAUTHORIZED, "访问受限", "")
	PARAM_VALID_ERROR = gcode.New(SC_BAD_REQUEST, "参数校验失败", "")
	PARAMETER_ERROR   = gcode.New(SC_BAD_REQUEST, "请求参数有误", "")
	PARAM_MISS        = gcode.New(SC_BAD_REQUEST, "缺少必要的请求参数", "")
	UNSUPPORTED_ERROR = gcode.New(SC_UNSUPPORTED_ERROR, "不支持的操作", "")
	FAILURE           = gcode.New(SC_FAILURE, "服务器异常，请稍后再试", "")
	SERVICE_ERROR     = gcode.New(SC_SERVICE_ERROR, "服务器异常，请稍后再试", "")
	UPLOAD_ERROR      = gcode.New(SC_UPLOAD_ERROR, "上传失败", "")
	RETRY_ERROR       = gcode.New(10011, "超过了最大重试次数[%d 次]，不允许重试!", "")
)
