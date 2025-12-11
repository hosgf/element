package uerrors

import (
	"fmt"

	"github.com/hosgf/element/model/result"
)

// 预定义的业务错误创建函数

// NewValidationError 创建参数验证错误
func NewValidationError(field, message string) *BizError {
	return NewBizError(
		ErrorTypeValidation,
		ErrorLevelWarning,
		int(result.REQ_REJECT.Code()),
		fmt.Sprintf("参数验证失败: %s", field),
		message,
	)
}

// NewBizLogicError 创建业务逻辑错误
func NewBizLogicError(code int, message string, details ...string) *BizError {
	return NewBizError(
		ErrorTypeBusiness,
		ErrorLevelWarning,
		code,
		message,
		details...,
	)
}

// NewSystemError 创建系统错误
func NewSystemError(message string, details ...string) *BizError {
	return NewBizError(
		ErrorTypeSystem,
		ErrorLevelError,
		result.SC_FAILURE,
		message,
		details...,
	)
}

// NewNetworkError 创建网络错误
func NewNetworkError(message string, details ...string) *BizError {
	return NewBizError(
		ErrorTypeNetwork,
		ErrorLevelError,
		result.SC_FAILURE,
		message,
		details...,
	)
}

// NewDatabaseError 创建数据库错误
func NewDatabaseError(message string, details ...string) *BizError {
	return NewBizError(
		ErrorTypeDatabase,
		ErrorLevelError,
		result.SC_FAILURE,
		message,
		details...,
	)
}

// NewExternalServiceError 创建外部服务错误
func NewExternalServiceError(serviceName, message string, details ...string) *BizError {
	return NewBizError(
		ErrorTypeExternal,
		ErrorLevelError,
		result.SC_FAILURE,
		fmt.Sprintf("外部服务 %s 错误: %s", serviceName, message),
		details...,
	)
}

// WrapSystemError 包装系统错误
func WrapSystemError(err error, message string) *BizError {
	return WrapError(err, ErrorTypeSystem, ErrorLevelError, result.SC_FAILURE, message)
}

// WrapNetworkError 包装网络错误
func WrapNetworkError(err error, message string) *BizError {
	return WrapError(err, ErrorTypeNetwork, ErrorLevelError, result.SC_FAILURE, message)
}

// WrapDatabaseError 包装数据库错误
func WrapDatabaseError(err error, message string) *BizError {
	return WrapError(err, ErrorTypeDatabase, ErrorLevelError, result.SC_FAILURE, message)
}

// WrapExternalServiceError 包装外部服务错误
func WrapExternalServiceError(err error, serviceName string) *BizError {
	message := fmt.Sprintf("外部服务 %s 调用失败", serviceName)
	return WrapError(err, ErrorTypeExternal, ErrorLevelError, result.SC_FAILURE, message)
}

// 常用业务错误码
const (
	// 用户相关错误
	CodeUserNotFound     = 1001
	CodeUserAlreadyExist = 1002
	CodeUserDisabled     = 1003
	CodeInvalidPassword  = 1004

	// 权限相关错误
	CodeUnauthorized = 2001
	CodeForbidden    = 2002
	CodeTokenExpired = 2003
	CodeInvalidToken = 2004

	// 业务逻辑错误
	CodeResourceNotFound = 3001
	CodeResourceConflict = 3002
	CodeOperationFailed  = 3003
	CodeQuotaExceeded    = 3004

	// 系统错误
	CodeInternalError      = 5001
	CodeServiceUnavailable = 5002
	CodeTimeout            = 5003
)

// 常用错误创建函数

// UserNotFound 用户不存在
func UserNotFound(userID string) *BizError {
	return NewBizLogicError(
		CodeUserNotFound,
		"用户不存在",
		fmt.Sprintf("用户ID: %s", userID),
	)
}

// UserAlreadyExist 用户已存在
func UserAlreadyExist(username string) *BizError {
	return NewBizLogicError(
		CodeUserAlreadyExist,
		"用户已存在",
		fmt.Sprintf("用户名: %s", username),
	)
}

// Unauthorized 未授权
func Unauthorized(message string) *BizError {
	return NewBizLogicError(
		CodeUnauthorized,
		"未授权访问",
		message,
	)
}

// Forbidden 禁止访问
func Forbidden(message string) *BizError {
	return NewBizLogicError(
		CodeForbidden,
		"禁止访问",
		message,
	)
}

// ResourceNotFound 资源不存在
func ResourceNotFound(resourceType, resourceID string) *BizError {
	return NewBizLogicError(
		CodeResourceNotFound,
		"资源不存在",
		fmt.Sprintf("资源类型: %s, 资源ID: %s", resourceType, resourceID),
	)
}

// ResourceConflict 资源冲突
func ResourceConflict(resourceType, message string) *BizError {
	return NewBizLogicError(
		CodeResourceConflict,
		"资源冲突",
		fmt.Sprintf("资源类型: %s, 原因: %s", resourceType, message),
	)
}

// OperationFailed 操作失败
func OperationFailed(operation, reason string) *BizError {
	return NewBizLogicError(
		CodeOperationFailed,
		"操作失败",
		fmt.Sprintf("操作: %s, 原因: %s", operation, reason),
	)
}

// QuotaExceeded 配额超限
func QuotaExceeded(quotaType string, current, limit int64) *BizError {
	return NewBizLogicError(
		CodeQuotaExceeded,
		"配额超限",
		fmt.Sprintf("配额类型: %s, 当前: %d, 限制: %d", quotaType, current, limit),
	)
}

// ServiceUnavailable 服务不可用
func ServiceUnavailable(serviceName string) *BizError {
	return NewSystemError(
		"服务不可用",
		fmt.Sprintf("服务: %s", serviceName),
	)
}

// Timeout 超时
func Timeout(operation string, timeoutSeconds int) *BizError {
	return NewSystemError(
		"操作超时",
		fmt.Sprintf("操作: %s, 超时时间: %d秒", operation, timeoutSeconds),
	)
}
