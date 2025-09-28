package proxy

import (
	"sync"
	"time"
)

// ============================================================================
// 熔断器管理
// ============================================================================

// CircuitBreaker 熔断器结构
type CircuitBreaker struct {
	FailureCount     int
	SuccessCount     int
	LastFailureTime  time.Time
	State            CircuitState
	FailureThreshold int
	RecoveryTimeout  time.Duration
	mutex            sync.RWMutex
}

// CircuitState 熔断器状态
type CircuitState int

const (
	StateClosed CircuitState = iota
	StateOpen
	StateHalfOpen
)

// 熔断器状态字符串
func (s CircuitState) String() string {
	switch s {
	case StateClosed:
		return "CLOSED"
	case StateOpen:
		return "OPEN"
	case StateHalfOpen:
		return "HALF_OPEN"
	default:
		return "UNKNOWN"
	}
}

// 全局熔断器状态
var (
	circuitBreakers = make(map[string]*CircuitBreaker)
	cbMutex         = &sync.RWMutex{}
)

// 检查熔断器是否开启
func (gw *Gateway) isCircuitBreakerOpen(routeKey string) bool {
	cbMutex.RLock()
	defer cbMutex.RUnlock()

	cb, exists := circuitBreakers[routeKey]
	if !exists {
		return false
	}

	cb.mutex.RLock()
	defer cb.mutex.RUnlock()

	switch cb.State {
	case StateOpen:
		// 检查是否可以进入半开状态
		if time.Since(cb.LastFailureTime) > cb.RecoveryTimeout {
			cb.State = StateHalfOpen
			return false
		}
		return true
	case StateHalfOpen:
		return false
	default:
		return false
	}
}

// 更新熔断器状态
func (gw *Gateway) updateCircuitBreaker(routeKey string, success bool) {
	cbMutex.Lock()
	defer cbMutex.Unlock()

	cb, exists := circuitBreakers[routeKey]
	if !exists {
		config := gw.getConfig()
		cbConfig := config.GetCircuitBreakerConfig()
		cb = &CircuitBreaker{
			FailureThreshold: cbConfig.FailureThreshold,
			RecoveryTimeout:  cbConfig.RecoveryTimeout,
			State:            StateClosed,
		}
		circuitBreakers[routeKey] = cb
	}

	cb.mutex.Lock()
	defer cb.mutex.Unlock()

	if success {
		cb.SuccessCount++
		if cb.State == StateHalfOpen {
			// 半开状态下成功，重置为关闭状态
			cb.State = StateClosed
			cb.FailureCount = 0
		}
	} else {
		cb.FailureCount++
		cb.LastFailureTime = time.Now()

		if cb.FailureCount >= cb.FailureThreshold {
			cb.State = StateOpen
		}
	}
}

// GetCircuitBreakerStatus 获取熔断器状态
func (gw *Gateway) GetCircuitBreakerStatus(routeKey string) *CircuitBreakerStatus {
	cbMutex.RLock()
	defer cbMutex.RUnlock()

	cb, exists := circuitBreakers[routeKey]
	if !exists {
		return &CircuitBreakerStatus{
			RouteKey: routeKey,
			State:    StateClosed.String(),
			Exists:   false,
		}
	}

	cb.mutex.RLock()
	defer cb.mutex.RUnlock()

	return &CircuitBreakerStatus{
		RouteKey:        routeKey,
		State:           cb.State.String(),
		FailureCount:    cb.FailureCount,
		SuccessCount:    cb.SuccessCount,
		LastFailureTime: cb.LastFailureTime,
		Exists:          true,
	}
}

// GetAllCircuitBreakerStatus 获取所有熔断器状态
func (gw *Gateway) GetAllCircuitBreakerStatus() map[string]*CircuitBreakerStatus {
	cbMutex.RLock()
	defer cbMutex.RUnlock()

	statuses := make(map[string]*CircuitBreakerStatus)
	for routeKey, cb := range circuitBreakers {
		cb.mutex.RLock()
		statuses[routeKey] = &CircuitBreakerStatus{
			RouteKey:        routeKey,
			State:           cb.State.String(),
			FailureCount:    cb.FailureCount,
			SuccessCount:    cb.SuccessCount,
			LastFailureTime: cb.LastFailureTime,
			Exists:          true,
		}
		cb.mutex.RUnlock()
	}
	return statuses
}

// ResetCircuitBreaker 重置熔断器
func (gw *Gateway) ResetCircuitBreaker(routeKey string) bool {
	cbMutex.Lock()
	defer cbMutex.Unlock()

	cb, exists := circuitBreakers[routeKey]
	if !exists {
		return false
	}

	cb.mutex.Lock()
	defer cb.mutex.Unlock()

	cb.State = StateClosed
	cb.FailureCount = 0
	cb.SuccessCount = 0
	cb.LastFailureTime = time.Time{}

	return true
}

// ResetAllCircuitBreakers 重置所有熔断器
func (gw *Gateway) ResetAllCircuitBreakers() {
	cbMutex.Lock()
	defer cbMutex.Unlock()

	for _, cb := range circuitBreakers {
		cb.mutex.Lock()
		cb.State = StateClosed
		cb.FailureCount = 0
		cb.SuccessCount = 0
		cb.LastFailureTime = time.Time{}
		cb.mutex.Unlock()
	}
}

// CircuitBreakerStatus 熔断器状态结构
type CircuitBreakerStatus struct {
	RouteKey        string    `json:"route_key"`
	State           string    `json:"state"`
	FailureCount    int       `json:"failure_count"`
	SuccessCount    int       `json:"success_count"`
	LastFailureTime time.Time `json:"last_failure_time"`
	Exists          bool      `json:"exists"`
}

// CircuitBreakerStats 熔断器统计
type CircuitBreakerStats struct {
	TotalBreakers    int               `json:"total_breakers"`
	OpenBreakers     int               `json:"open_breakers"`
	HalfOpenBreakers int               `json:"half_open_breakers"`
	ClosedBreakers   int               `json:"closed_breakers"`
	TotalFailures    int               `json:"total_failures"`
	TotalSuccesses   int               `json:"total_successes"`
	Breakers         map[string]string `json:"breakers"`
}

// GetCircuitBreakerStats 获取熔断器统计
func (gw *Gateway) GetCircuitBreakerStats() *CircuitBreakerStats {
	cbMutex.RLock()
	defer cbMutex.RUnlock()

	stats := &CircuitBreakerStats{
		TotalBreakers: len(circuitBreakers),
		Breakers:      make(map[string]string),
	}

	for routeKey, cb := range circuitBreakers {
		cb.mutex.RLock()
		state := cb.State.String()
		stats.Breakers[routeKey] = state

		switch cb.State {
		case StateOpen:
			stats.OpenBreakers++
		case StateHalfOpen:
			stats.HalfOpenBreakers++
		case StateClosed:
			stats.ClosedBreakers++
		}

		stats.TotalFailures += cb.FailureCount
		stats.TotalSuccesses += cb.SuccessCount
		cb.mutex.RUnlock()
	}

	return stats
}
