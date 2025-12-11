package uerrors

import (
	"context"
	"fmt"

	"github.com/hosgf/element/model/result"
)

// PanicIf 在条件成立时抛出业务错误（panic）
func PanicIf(condition bool, err *BizError) {
	if condition {
		panic(err)
	}
}

// Must 对返回(err)的函数进行检查，若err不为空则panic包装为系统错误
func Must(err error, message string) {
	if err != nil {
		panic(WrapError(err, ErrorTypeSystem, ErrorLevelError, result.SC_FAILURE, message))
	}
}

// MustWithContext 对返回(err)的函数进行检查，若err不为空则panic包装为系统错误，并添加上下文信息
func MustWithContext(ctx context.Context, err error, message string) {
	if err != nil {
		bizErr := WrapError(err, ErrorTypeSystem, ErrorLevelError, result.SC_FAILURE, message)
		if ctx != nil {
			bizErr.RequestID = GetRequestID(ctx)
			bizErr.UserID = GetUserID(ctx)
		}
		panic(bizErr)
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

// WrapKubernetesError 包装Kubernetes API错误
func WrapKubernetesError(ctx context.Context, err error, operation string) *BizError {
	if err == nil {
		return nil
	}

	bizErr := WrapError(err, ErrorTypeExternal, ErrorLevelError, result.SC_FAILURE,
		fmt.Sprintf("Kubernetes操作失败: %s", operation))

	if ctx != nil {
		bizErr.RequestID = GetRequestID(ctx)
		bizErr.UserID = GetUserID(ctx)
	}

	return bizErr
}

// NewKubernetesError 创建Kubernetes错误
func NewKubernetesError(ctx context.Context, operation, message string, details ...string) *BizError {
	bizErr := NewSystemError(
		fmt.Sprintf("Kubernetes操作失败: %s - %s", operation, message),
		details...,
	)

	if ctx != nil {
		bizErr.RequestID = GetRequestID(ctx)
		bizErr.UserID = GetUserID(ctx)
	}

	return bizErr
}
