package uerrors

import (
	"errors"
	"fmt"
	"net/http"
	"runtime"
	"strings"
	"time"

	"github.com/hosgf/element/model/result"
)

// ErrorType 错误类型枚举
type ErrorType int

const (
	// ErrorTypeUnknown 未知错误
	ErrorTypeUnknown ErrorType = iota
	// ErrorTypeValidation 参数验证错误
	ErrorTypeValidation
	// ErrorTypeBusiness 业务逻辑错误
	ErrorTypeBusiness
	// ErrorTypeSystem 系统错误
	ErrorTypeSystem
	// ErrorTypeNetwork 网络错误
	ErrorTypeNetwork
	// ErrorTypeDatabase 数据库错误
	ErrorTypeDatabase
	// ErrorTypeExternal 外部服务错误
	ErrorTypeExternal
)

// ErrorLevel 错误级别
type ErrorLevel int

const (
	// ErrorLevelInfo 信息级别
	ErrorLevelInfo ErrorLevel = iota
	// ErrorLevelWarning 警告级别
	ErrorLevelWarning
	// ErrorLevelError 错误级别
	ErrorLevelError
	// ErrorLevelCritical 严重错误级别
	ErrorLevelCritical
)

// BizError 业务错误结构
type BizError struct {
	Type        ErrorType   `json:"type"`
	Level       ErrorLevel  `json:"level"`
	Code        int         `json:"code"`
	Message     string      `json:"message"`
	Details     string      `json:"details,omitempty"`
	Stack       string      `json:"stack,omitempty"`
	Timestamp   time.Time   `json:"timestamp"`
	RequestID   string      `json:"request_id,omitempty"`
	UserID      string      `json:"user_id,omitempty"`
	Context     interface{} `json:"context,omitempty"`
	OriginalErr error       `json:"-"`
}

// Error 实现error接口
func (e *BizError) Error() string {
	if e.Details != "" {
		return fmt.Sprintf("[%s] %s: %s", e.getTypeString(), e.Message, e.Details)
	}
	return fmt.Sprintf("[%s] %s", e.getTypeString(), e.Message)
}

// Unwrap 支持错误链
func (e *BizError) Unwrap() error {
	return e.OriginalErr
}

// TypeString 导出错误类型字符串
func (e *BizError) TypeString() string {
	return e.getTypeString()
}

// LevelString 导出错误级别字符串
func (e *BizError) LevelString() string {
	return e.getLevelString()
}

// getTypeString 获取错误类型字符串
func (e *BizError) getTypeString() string {
	switch e.Type {
	case ErrorTypeValidation:
		return "VALIDATION"
	case ErrorTypeBusiness:
		return "BUSINESS"
	case ErrorTypeSystem:
		return "SYSTEM"
	case ErrorTypeNetwork:
		return "NETWORK"
	case ErrorTypeDatabase:
		return "DATABASE"
	case ErrorTypeExternal:
		return "EXTERNAL"
	default:
		return "UNKNOWN"
	}
}

// getLevelString 获取错误级别字符串
func (e *BizError) getLevelString() string {
	switch e.Level {
	case ErrorLevelInfo:
		return "INFO"
	case ErrorLevelWarning:
		return "WARNING"
	case ErrorLevelError:
		return "ERROR"
	case ErrorLevelCritical:
		return "CRITICAL"
	default:
		return "UNKNOWN"
	}
}

// ToResultResponse 转换为element包的响应格式
func (e *BizError) ToResultResponse() *result.Response {
	response := result.NewResponse()
	response.Code = e.Code
	response.Message = e.Message
	return response
}

// GetHTTPStatus 获取对应的HTTP状态码
func (e *BizError) GetHTTPStatus() int {
	switch e.Type {
	case ErrorTypeValidation:
		return http.StatusBadRequest
	case ErrorTypeBusiness:
		return http.StatusUnprocessableEntity
	case ErrorTypeSystem:
		return http.StatusInternalServerError
	case ErrorTypeNetwork:
		return http.StatusBadGateway
	case ErrorTypeDatabase:
		return http.StatusInternalServerError
	case ErrorTypeExternal:
		return http.StatusBadGateway
	default:
		return http.StatusInternalServerError
	}
}

// NewBizError 创建业务错误
func NewBizError(errType ErrorType, level ErrorLevel, code int, message string, details ...string) *BizError {
	stack := getStackTrace()
	detail := ""
	if len(details) > 0 {
		detail = details[0]
	}

	return &BizError{
		Type:      errType,
		Level:     level,
		Code:      code,
		Message:   message,
		Details:   detail,
		Stack:     stack,
		Timestamp: time.Now(),
	}
}

// WrapError 包装现有错误
func WrapError(err error, errType ErrorType, level ErrorLevel, code int, message string) *BizError {
	stack := getStackTrace()

	return &BizError{
		Type:        errType,
		Level:       level,
		Code:        code,
		Message:     message,
		Stack:       stack,
		Timestamp:   time.Now(),
		OriginalErr: err,
	}
}

// getStackTrace 获取堆栈跟踪
func getStackTrace() string {
	const depth = 32
	var pcs [depth]uintptr
	n := runtime.Callers(3, pcs[:])

	var builder strings.Builder
	frames := runtime.CallersFrames(pcs[:n])

	for {
		frame, more := frames.Next()
		if !more {
			break
		}
		builder.WriteString(fmt.Sprintf("%s:%d %s\n", frame.File, frame.Line, frame.Function))
	}

	return builder.String()
}

// IsBizError 检查是否为业务错误
func IsBizError(err error) (*BizError, bool) {
	var bizErr *BizError
	if errors.As(err, &bizErr) {
		return bizErr, true
	}
	return nil, false
}
