package proxy

import (
	"context"
	"errors"
	"net"
	"strings"
	"time"

	"github.com/hosgf/element/logger"
)

// ============================================================================
// 错误处理机制
// ============================================================================

const (
	SC_BAD_REQUEST   = 400
	SC_NOT_FOUND     = 404
	SC_BAD_GATEWAY   = 502
	SC_FAILURE       = 500
	SC_GATEWAY       = 4001
	SC_TIMEOUT       = 4008
	SC_SERVICE_ERROR = 5700
)

// ErrorType 错误类型
type ErrorType int

const (
	ErrorTypeNetwork ErrorType = iota
	ErrorTypeTimeout
	ErrorTypeCircuitBreaker
	ErrorTypeValidation
	ErrorTypeInternal
	ErrorTypeUnknown
)

// 错误类型字符串
func (et ErrorType) String() string {
	switch et {
	case ErrorTypeNetwork:
		return "NETWORK"
	case ErrorTypeTimeout:
		return "TIMEOUT"
	case ErrorTypeCircuitBreaker:
		return "CIRCUIT_BREAKER"
	case ErrorTypeValidation:
		return "VALIDATION"
	case ErrorTypeInternal:
		return "INTERNAL"
	default:
		return "UNKNOWN"
	}
}

// ProxyError 代理错误结构
type ProxyError struct {
	Type      ErrorType
	Code      int
	Message   string
	Details   string
	Timestamp time.Time
	Context   map[string]interface{}
}

// 实现error接口
func (e *ProxyError) Error() string {
	return e.Message
}

// NewNetworkError 创建网络错误
func NewNetworkError(message, details string) *ProxyError {
	return &ProxyError{
		Type:      ErrorTypeNetwork,
		Code:      SC_GATEWAY,
		Message:   message,
		Details:   details,
		Timestamp: time.Now(),
		Context:   make(map[string]interface{}),
	}
}

// NewTimeoutError 创建超时错误
func NewTimeoutError(message, details string) *ProxyError {
	return &ProxyError{
		Type:      ErrorTypeTimeout,
		Code:      SC_TIMEOUT,
		Message:   message,
		Details:   details,
		Timestamp: time.Now(),
		Context:   make(map[string]interface{}),
	}
}

// NewCircuitBreakerError 创建熔断器错误
func NewCircuitBreakerError(message, details string) *ProxyError {
	return &ProxyError{
		Type:      ErrorTypeCircuitBreaker,
		Code:      SC_SERVICE_ERROR,
		Message:   message,
		Details:   details,
		Timestamp: time.Now(),
		Context:   make(map[string]interface{}),
	}
}

// NewValidationError 创建验证错误
func NewValidationError(message, details string) *ProxyError {
	return &ProxyError{
		Type:      ErrorTypeValidation,
		Code:      SC_BAD_REQUEST,
		Message:   message,
		Details:   details,
		Timestamp: time.Now(),
		Context:   make(map[string]interface{}),
	}
}

// NewInternalError 创建内部错误
func NewInternalError(message, details string) *ProxyError {
	return &ProxyError{
		Type:      ErrorTypeInternal,
		Code:      SC_FAILURE,
		Message:   message,
		Details:   details,
		Timestamp: time.Now(),
		Context:   make(map[string]interface{}),
	}
}

// ErrorHandler 错误处理器
type ErrorHandler struct {
	// 错误统计
	errorStats map[ErrorType]int64
	// 错误日志记录
	logErrors bool
	// 错误上下文
	context map[string]interface{}
}

// NewErrorHandler 创建新的错误处理器
func NewErrorHandler() *ErrorHandler {
	return &ErrorHandler{
		errorStats: make(map[ErrorType]int64),
		logErrors:  true,
		context:    make(map[string]interface{}),
	}
}

// 全局错误处理器
var errorHandler = NewErrorHandler()

// HandleError 处理错误
func (eh *ErrorHandler) HandleError(ctx context.Context, err error, target string) *ProxyError {
	var proxyErr *ProxyError

	// 检查是否已经是ProxyError
	if pe, ok := err.(*ProxyError); ok {
		proxyErr = pe
	} else {
		// 分类错误
		proxyErr = eh.classifyError(err, target)
	}

	// 记录错误统计
	eh.errorStats[proxyErr.Type]++

	// 记录错误日志
	if eh.logErrors {
		eh.logError(ctx, proxyErr, target)
	}

	// 添加上下文信息
	proxyErr.Context["target"] = target
	proxyErr.Context["timestamp"] = proxyErr.Timestamp

	return proxyErr
}

// 分类错误
func (eh *ErrorHandler) classifyError(err error, target string) *ProxyError {
	// DNS/连接错误
	var ne *net.OpError
	if errors.As(err, &ne) {
		return NewNetworkError("服务连接失败，请稍后重试", err.Error())
	}

	// 超时错误
	if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
		return NewTimeoutError("请求超时，请稍后再试", err.Error())
	}

	// 熔断器错误
	if strings.Contains(err.Error(), "熔断器已开启") {
		return NewCircuitBreakerError("服务暂时不可用，请稍后重试", err.Error())
	}

	// 其他错误
	return NewInternalError("服务器内部错误", err.Error())
}

// 记录错误日志
func (eh *ErrorHandler) logError(ctx context.Context, proxyErr *ProxyError, target string) {
	switch proxyErr.Type {
	case ErrorTypeNetwork:
		logger.Errorf(ctx, "network error: %v, target=%s, details=%s", proxyErr.Message, target, proxyErr.Details)
	case ErrorTypeTimeout:
		logger.Errorf(ctx, "timeout error: %v, target=%s, details=%s", proxyErr.Message, target, proxyErr.Details)
	case ErrorTypeCircuitBreaker:
		logger.Errorf(ctx, "circuit breaker error: %v, target=%s, details=%s", proxyErr.Message, target, proxyErr.Details)
	case ErrorTypeValidation:
		logger.Errorf(ctx, "validation error: %v, target=%s, details=%s", proxyErr.Message, target, proxyErr.Details)
	case ErrorTypeInternal:
		logger.Errorf(ctx, "internal error: %v, target=%s, details=%s", proxyErr.Message, target, proxyErr.Details)
	default:
		logger.Errorf(ctx, "unknown error: %v, target=%s, details=%s", proxyErr.Message, target, proxyErr.Details)
	}
}

// GetErrorStats 获取错误统计
func (eh *ErrorHandler) GetErrorStats() map[string]int64 {
	stats := make(map[string]int64)
	for errorType, count := range eh.errorStats {
		stats[errorType.String()] = count
	}
	return stats
}

// ResetErrorStats 重置错误统计
func (eh *ErrorHandler) ResetErrorStats() {
	eh.errorStats = make(map[ErrorType]int64)
}

// SetLogErrors 设置错误日志记录
func (eh *ErrorHandler) SetLogErrors(logErrors bool) {
	eh.logErrors = logErrors
}

// AddContext 添加错误上下文
func (eh *ErrorHandler) AddContext(key string, value interface{}) {
	eh.context[key] = value
}

// GetContext 获取错误上下文
func (eh *ErrorHandler) GetContext() map[string]interface{} {
	return eh.context
}

// ErrorRecovery 错误恢复机制
type ErrorRecovery struct {
	// 重试次数
	maxRetries int
	// 重试间隔
	retryInterval time.Duration
	// 可恢复的错误类型
	recoverableErrors []ErrorType
}

// NewErrorRecovery 创建新的错误恢复机制
func NewErrorRecovery(maxRetries int, retryInterval time.Duration) *ErrorRecovery {
	return &ErrorRecovery{
		maxRetries:    maxRetries,
		retryInterval: retryInterval,
		recoverableErrors: []ErrorType{
			ErrorTypeNetwork,
			ErrorTypeTimeout,
		},
	}
}

// IsRecoverable 检查错误是否可恢复
func (er *ErrorRecovery) IsRecoverable(proxyErr *ProxyError) bool {
	for _, errorType := range er.recoverableErrors {
		if proxyErr.Type == errorType {
			return true
		}
	}
	return false
}

// GetMaxRetries 获取重试次数
func (er *ErrorRecovery) GetMaxRetries() int {
	return er.maxRetries
}

// GetRetryInterval 获取重试间隔
func (er *ErrorRecovery) GetRetryInterval() time.Duration {
	return er.retryInterval
}

// 全局错误恢复机制
var errorRecovery = NewErrorRecovery(3, time.Second)

// HandleError 处理错误（Gateway方法）
func (gw *Gateway) HandleError(ctx context.Context, err error, target string) *ProxyError {
	return errorHandler.HandleError(ctx, err, target)
}

// GetErrorStats 获取错误统计（Gateway方法）
func (gw *Gateway) GetErrorStats() map[string]int64 {
	return errorHandler.GetErrorStats()
}

// ResetErrorStats 重置错误统计（Gateway方法）
func (gw *Gateway) ResetErrorStats() {
	errorHandler.ResetErrorStats()
}
