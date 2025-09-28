package proxy

import (
	"sync"
	"time"
)

// ============================================================================
// 配置管理
// ============================================================================

// GatewayConfig 网关配置结构
type GatewayConfig struct {
	// HTTP客户端配置
	Timeout               time.Duration `json:"timeout" yaml:"timeout"`
	MaxIdleConns          int           `json:"max_idle_conns" yaml:"max_idle_conns"`
	MaxIdleConnsPerHost   int           `json:"max_idle_conns_per_host" yaml:"max_idle_conns_per_host"`
	MaxConnsPerHost       int           `json:"max_conns_per_host" yaml:"max_conns_per_host"`
	IdleConnTimeout       time.Duration `json:"idle_conn_timeout" yaml:"idle_conn_timeout"`
	TLSHandshakeTimeout   time.Duration `json:"tls_handshake_timeout" yaml:"tls_handshake_timeout"`
	ExpectContinueTimeout time.Duration `json:"expect_continue_timeout" yaml:"expect_continue_timeout"`

	// 重试配置
	MaxRetries int `json:"max_retries" yaml:"max_retries"`

	// 熔断器配置
	FailureThreshold int           `json:"failure_threshold" yaml:"failure_threshold"`
	RecoveryTimeout  time.Duration `json:"recovery_timeout" yaml:"recovery_timeout"`
}

// DefaultGatewayConfig 默认配置
func DefaultGatewayConfig() *GatewayConfig {
	return &GatewayConfig{
		Timeout:               30 * time.Second,
		MaxIdleConns:          1000,
		MaxIdleConnsPerHost:   100,
		MaxConnsPerHost:       200,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		MaxRetries:            3,
		FailureThreshold:      5,
		RecoveryTimeout:       30 * time.Second,
	}
}

// 全局配置变量
var (
	globalConfig = DefaultGatewayConfig()
	configMutex  = &sync.RWMutex{}
)

// 获取配置
func getConfig() *GatewayConfig {
	configMutex.RLock()
	defer configMutex.RUnlock()
	return globalConfig
}

// SetConfig 设置配置
func SetConfig(config *GatewayConfig) {
	configMutex.Lock()
	defer configMutex.Unlock()
	globalConfig = config
}

// 获取配置（Gateway方法）
func (gw *Gateway) getConfig() *GatewayConfig {
	return getConfig()
}

// Validate 验证配置
func (config *GatewayConfig) Validate() error {
	if config.Timeout <= 0 {
		return &ConfigError{Field: "timeout", Message: "timeout must be positive"}
	}
	if config.MaxIdleConns <= 0 {
		return &ConfigError{Field: "max_idle_conns", Message: "max_idle_conns must be positive"}
	}
	if config.MaxIdleConnsPerHost <= 0 {
		return &ConfigError{Field: "max_idle_conns_per_host", Message: "max_idle_conns_per_host must be positive"}
	}
	if config.MaxConnsPerHost <= 0 {
		return &ConfigError{Field: "max_conns_per_host", Message: "max_conns_per_host must be positive"}
	}
	if config.IdleConnTimeout <= 0 {
		return &ConfigError{Field: "idle_conn_timeout", Message: "idle_conn_timeout must be positive"}
	}
	if config.TLSHandshakeTimeout <= 0 {
		return &ConfigError{Field: "tls_handshake_timeout", Message: "tls_handshake_timeout must be positive"}
	}
	if config.ExpectContinueTimeout <= 0 {
		return &ConfigError{Field: "expect_continue_timeout", Message: "expect_continue_timeout must be positive"}
	}
	if config.MaxRetries < 0 {
		return &ConfigError{Field: "max_retries", Message: "max_retries must be non-negative"}
	}
	if config.FailureThreshold <= 0 {
		return &ConfigError{Field: "failure_threshold", Message: "failure_threshold must be positive"}
	}
	if config.RecoveryTimeout <= 0 {
		return &ConfigError{Field: "recovery_timeout", Message: "recovery_timeout must be positive"}
	}
	return nil
}

// ConfigError 配置错误结构
type ConfigError struct {
	Field   string
	Message string
}

func (e *ConfigError) Error() string {
	return "config error: " + e.Field + " - " + e.Message
}

// Merge 合并配置（用于部分更新）
func (config *GatewayConfig) Merge(other *GatewayConfig) *GatewayConfig {
	if other == nil {
		return config
	}

	merged := &GatewayConfig{
		Timeout:               config.Timeout,
		MaxIdleConns:          config.MaxIdleConns,
		MaxIdleConnsPerHost:   config.MaxIdleConnsPerHost,
		MaxConnsPerHost:       config.MaxConnsPerHost,
		IdleConnTimeout:       config.IdleConnTimeout,
		TLSHandshakeTimeout:   config.TLSHandshakeTimeout,
		ExpectContinueTimeout: config.ExpectContinueTimeout,
		MaxRetries:            config.MaxRetries,
		FailureThreshold:      config.FailureThreshold,
		RecoveryTimeout:       config.RecoveryTimeout,
	}

	if other.Timeout > 0 {
		merged.Timeout = other.Timeout
	}
	if other.MaxIdleConns > 0 {
		merged.MaxIdleConns = other.MaxIdleConns
	}
	if other.MaxIdleConnsPerHost > 0 {
		merged.MaxIdleConnsPerHost = other.MaxIdleConnsPerHost
	}
	if other.MaxConnsPerHost > 0 {
		merged.MaxConnsPerHost = other.MaxConnsPerHost
	}
	if other.IdleConnTimeout > 0 {
		merged.IdleConnTimeout = other.IdleConnTimeout
	}
	if other.TLSHandshakeTimeout > 0 {
		merged.TLSHandshakeTimeout = other.TLSHandshakeTimeout
	}
	if other.ExpectContinueTimeout > 0 {
		merged.ExpectContinueTimeout = other.ExpectContinueTimeout
	}
	if other.MaxRetries >= 0 {
		merged.MaxRetries = other.MaxRetries
	}
	if other.FailureThreshold > 0 {
		merged.FailureThreshold = other.FailureThreshold
	}
	if other.RecoveryTimeout > 0 {
		merged.RecoveryTimeout = other.RecoveryTimeout
	}

	return merged
}

// GetHTTPClientConfig 获取HTTP客户端配置
func (config *GatewayConfig) GetHTTPClientConfig() HTTPClientConfig {
	return HTTPClientConfig{
		Timeout:               config.Timeout,
		MaxIdleConns:          config.MaxIdleConns,
		MaxIdleConnsPerHost:   config.MaxIdleConnsPerHost,
		MaxConnsPerHost:       config.MaxConnsPerHost,
		IdleConnTimeout:       config.IdleConnTimeout,
		TLSHandshakeTimeout:   config.TLSHandshakeTimeout,
		ExpectContinueTimeout: config.ExpectContinueTimeout,
	}
}

// GetRetryConfig 获取重试配置
func (config *GatewayConfig) GetRetryConfig() RetryConfig {
	return RetryConfig{
		MaxRetries: config.MaxRetries,
	}
}

// 获取熔断器配置
func (config *GatewayConfig) GetCircuitBreakerConfig() CircuitBreakerConfig {
	return CircuitBreakerConfig{
		FailureThreshold: config.FailureThreshold,
		RecoveryTimeout:  config.RecoveryTimeout,
	}
}

// HTTPClientConfig HTTP客户端配置
type HTTPClientConfig struct {
	Timeout               time.Duration
	MaxIdleConns          int
	MaxIdleConnsPerHost   int
	MaxConnsPerHost       int
	IdleConnTimeout       time.Duration
	TLSHandshakeTimeout   time.Duration
	ExpectContinueTimeout time.Duration
}

// RetryConfig 重试配置
type RetryConfig struct {
	MaxRetries int
}

// CircuitBreakerConfig 熔断器配置
type CircuitBreakerConfig struct {
	FailureThreshold int
	RecoveryTimeout  time.Duration
}
