package exception

import "fmt"

// PanicIf 在条件成立时抛出业务错误（panic）
func PanicIf(condition bool, err *BizError) {
	if condition {
		panic(err)
	}
}

// Must 对返回(err)的函数进行检查，若err不为空则panic包装为系统错误
func Must(err error, message string) {
	if err != nil {
		panic(WrapError(err, ErrorTypeSystem, ErrorLevelError, 500, message))
	}
}

// ThrowBiz 抛出业务错误
func ThrowBiz(code int, message string, details ...string) {
	panic(NewBizLogicError(code, message, details...))
}

// ThrowValidation 抛出校验错误
func ThrowValidation(field, message string) {
	panic(NewValidationError(field, message))
}

// ThrowSystem 抛出系统错误
func ThrowSystem(message string, details ...string) {
	panic(NewSystemError(message, details...))
}

// Throwf 使用格式化消息抛错（业务类）
func Throwf(code int, format string, a ...interface{}) {
	panic(NewBizLogicError(code, fmt.Sprintf(format, a...)))
}
