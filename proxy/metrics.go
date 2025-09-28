package proxy

import (
	"fmt"
	"sync"
	"time"
)

// ============================================================================
// 性能指标管理
// ============================================================================

// Metrics ProxyMetrics 性能指标结构
type Metrics struct {
	RequestCount            int64
	SuccessCount            int64
	FailureCount            int64
	TotalLatency            time.Duration
	CircuitBreakerOpenCount int64
	mutex                   *sync.RWMutex
}

// 全局性能指标实例
var (
	metrics = &Metrics{
		RequestCount:            0,
		SuccessCount:            0,
		FailureCount:            0,
		TotalLatency:            0,
		CircuitBreakerOpenCount: 0,
		mutex:                   &sync.RWMutex{},
	}
)

// GetMetrics 获取性能指标
func (gw *Gateway) GetMetrics() map[string]interface{} {
	metrics.mutex.RLock()
	defer metrics.mutex.RUnlock()

	avgLatency := time.Duration(0)
	if metrics.RequestCount > 0 {
		avgLatency = metrics.TotalLatency / time.Duration(metrics.RequestCount)
	}

	successRate := float64(0)
	if metrics.RequestCount > 0 {
		successRate = float64(metrics.SuccessCount) / float64(metrics.RequestCount) * 100
	}

	return map[string]interface{}{
		"request_count":              metrics.RequestCount,
		"success_count":              metrics.SuccessCount,
		"failure_count":              metrics.FailureCount,
		"success_rate":               fmt.Sprintf("%.2f%%", successRate),
		"average_latency":            avgLatency.String(),
		"circuit_breaker_open_count": metrics.CircuitBreakerOpenCount,
	}
}

// 记录请求指标
func (gw *Gateway) recordMetrics(success bool, latency time.Duration, circuitBreakerOpen bool) {
	metrics.mutex.Lock()
	defer metrics.mutex.Unlock()

	metrics.RequestCount++
	metrics.TotalLatency += latency

	if success {
		metrics.SuccessCount++
	} else {
		metrics.FailureCount++
	}

	if circuitBreakerOpen {
		metrics.CircuitBreakerOpenCount++
	}
}

// ResetMetrics 重置指标
func (gw *Gateway) ResetMetrics() {
	metrics.mutex.Lock()
	defer metrics.mutex.Unlock()

	metrics.RequestCount = 0
	metrics.SuccessCount = 0
	metrics.FailureCount = 0
	metrics.TotalLatency = 0
	metrics.CircuitBreakerOpenCount = 0
}

// GetDetailedMetrics 获取详细指标
func (gw *Gateway) GetDetailedMetrics() *DetailedMetrics {
	metrics.mutex.RLock()
	defer metrics.mutex.RUnlock()

	avgLatency := time.Duration(0)
	if metrics.RequestCount > 0 {
		avgLatency = metrics.TotalLatency / time.Duration(metrics.RequestCount)
	}

	successRate := float64(0)
	if metrics.RequestCount > 0 {
		successRate = float64(metrics.SuccessCount) / float64(metrics.RequestCount) * 100
	}

	failureRate := float64(0)
	if metrics.RequestCount > 0 {
		failureRate = float64(metrics.FailureCount) / float64(metrics.RequestCount) * 100
	}

	return &DetailedMetrics{
		RequestCount:            metrics.RequestCount,
		SuccessCount:            metrics.SuccessCount,
		FailureCount:            metrics.FailureCount,
		TotalLatency:            metrics.TotalLatency,
		AverageLatency:          avgLatency,
		SuccessRate:             successRate,
		FailureRate:             failureRate,
		CircuitBreakerOpenCount: metrics.CircuitBreakerOpenCount,
	}
}

// DetailedMetrics 详细指标结构
type DetailedMetrics struct {
	RequestCount            int64         `json:"request_count"`
	SuccessCount            int64         `json:"success_count"`
	FailureCount            int64         `json:"failure_count"`
	TotalLatency            time.Duration `json:"total_latency"`
	AverageLatency          time.Duration `json:"average_latency"`
	SuccessRate             float64       `json:"success_rate"`
	FailureRate             float64       `json:"failure_rate"`
	CircuitBreakerOpenCount int64         `json:"circuit_breaker_open_count"`
}

// MetricsSnapshot 指标快照
type MetricsSnapshot struct {
	Timestamp time.Time        `json:"timestamp"`
	Metrics   *DetailedMetrics `json:"metrics"`
}

// GetMetricsSnapshot 获取指标快照
func (gw *Gateway) GetMetricsSnapshot() *MetricsSnapshot {
	return &MetricsSnapshot{
		Timestamp: time.Now(),
		Metrics:   gw.GetDetailedMetrics(),
	}
}

// MetricsStats 指标统计
type MetricsStats struct {
	TotalRequests     int64         `json:"total_requests"`
	TotalSuccess      int64         `json:"total_success"`
	TotalFailures     int64         `json:"total_failures"`
	TotalLatency      time.Duration `json:"total_latency"`
	AverageLatency    time.Duration `json:"average_latency"`
	SuccessRate       float64       `json:"success_rate"`
	FailureRate       float64       `json:"failure_rate"`
	CircuitBreakerOps int64         `json:"circuit_breaker_operations"`
	Uptime            time.Duration `json:"uptime"`
}

// GetMetricsStats 获取指标统计
func (gw *Gateway) GetMetricsStats() *MetricsStats {
	metrics.mutex.RLock()
	defer metrics.mutex.RUnlock()

	avgLatency := time.Duration(0)
	if metrics.RequestCount > 0 {
		avgLatency = metrics.TotalLatency / time.Duration(metrics.RequestCount)
	}

	successRate := float64(0)
	if metrics.RequestCount > 0 {
		successRate = float64(metrics.SuccessCount) / float64(metrics.RequestCount) * 100
	}

	failureRate := float64(0)
	if metrics.RequestCount > 0 {
		failureRate = float64(metrics.FailureCount) / float64(metrics.RequestCount) * 100
	}

	return &MetricsStats{
		TotalRequests:     metrics.RequestCount,
		TotalSuccess:      metrics.SuccessCount,
		TotalFailures:     metrics.FailureCount,
		TotalLatency:      metrics.TotalLatency,
		AverageLatency:    avgLatency,
		SuccessRate:       successRate,
		FailureRate:       failureRate,
		CircuitBreakerOps: metrics.CircuitBreakerOpenCount,
		Uptime:            time.Since(startTime),
	}
}

// 启动时间
var startTime = time.Now()

// MetricsHealth 指标健康检查
type MetricsHealth struct {
	IsHealthy      bool          `json:"is_healthy"`
	SuccessRate    float64       `json:"success_rate"`
	AverageLatency time.Duration `json:"average_latency"`
	Message        string        `json:"message"`
}

// GetMetricsHealth 获取指标健康状态
func (gw *Gateway) GetMetricsHealth() *MetricsHealth {
	stats := gw.GetMetricsStats()

	// 健康检查标准
	isHealthy := true
	message := "指标正常"

	if stats.SuccessRate < 90.0 {
		isHealthy = false
		message = "成功率过低"
	}

	if stats.AverageLatency > 5*time.Second {
		isHealthy = false
		message = "平均延迟过高"
	}

	if stats.CircuitBreakerOps > 100 {
		isHealthy = false
		message = "熔断器触发次数过多"
	}

	return &MetricsHealth{
		IsHealthy:      isHealthy,
		SuccessRate:    stats.SuccessRate,
		AverageLatency: stats.AverageLatency,
		Message:        message,
	}
}

// ExportMetrics 指标导出
func (gw *Gateway) ExportMetrics() map[string]interface{} {
	return map[string]interface{}{
		"summary":  gw.GetMetrics(),
		"detailed": gw.GetDetailedMetrics(),
		"snapshot": gw.GetMetricsSnapshot(),
		"stats":    gw.GetMetricsStats(),
		"health":   gw.GetMetricsHealth(),
	}
}
