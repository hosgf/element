package proxy

import (
	"net/http"
	"sync"
	"time"
)

// ============================================================================
// HTTP客户端管理
// ============================================================================

// HTTPClientPool HTTP客户端池
type HTTPClientPool struct {
	pool *sync.Pool
}

// 全局HTTP客户端池
var (
	httpClientPool = &HTTPClientPool{
		pool: &sync.Pool{
			New: func() interface{} {
				config := getConfig()
				httpConfig := config.GetHTTPClientConfig()
				return &http.Client{
					Timeout: httpConfig.Timeout,
					Transport: &http.Transport{
						MaxIdleConns:          httpConfig.MaxIdleConns,
						MaxIdleConnsPerHost:   httpConfig.MaxIdleConnsPerHost,
						MaxConnsPerHost:       httpConfig.MaxConnsPerHost,
						IdleConnTimeout:       httpConfig.IdleConnTimeout,
						TLSHandshakeTimeout:   httpConfig.TLSHandshakeTimeout,
						ExpectContinueTimeout: httpConfig.ExpectContinueTimeout,
						DisableKeepAlives:     false,
						DisableCompression:    false,
					},
				}
			},
		},
	}
)

// 从连接池获取HTTP客户端
func getHTTPClient() *http.Client {
	return httpClientPool.pool.Get().(*http.Client)
}

// 归还HTTP客户端到连接池
func putHTTPClient(client *http.Client) {
	httpClientPool.pool.Put(client)
}

// 创建自定义HTTP客户端
func createHTTPClient(config *HTTPClientConfig) *http.Client {
	return &http.Client{
		Timeout: config.Timeout,
		Transport: &http.Transport{
			MaxIdleConns:          config.MaxIdleConns,
			MaxIdleConnsPerHost:   config.MaxIdleConnsPerHost,
			MaxConnsPerHost:       config.MaxConnsPerHost,
			IdleConnTimeout:       config.IdleConnTimeout,
			TLSHandshakeTimeout:   config.TLSHandshakeTimeout,
			ExpectContinueTimeout: config.ExpectContinueTimeout,
			DisableKeepAlives:     false,
			DisableCompression:    false,
		},
	}
}

// HTTPClientStats HTTP客户端统计
type HTTPClientStats struct {
	PoolSize       int           `json:"pool_size"`
	ActiveClients  int           `json:"active_clients"`
	IdleClients    int           `json:"idle_clients"`
	TotalRequests  int64         `json:"total_requests"`
	TotalErrors    int64         `json:"total_errors"`
	AverageLatency time.Duration `json:"average_latency"`
	LastUsed       time.Time     `json:"last_used"`
}

// GetHTTPClientStats 获取HTTP客户端统计
func (gw *Gateway) GetHTTPClientStats() *HTTPClientStats {
	// 这里可以添加更详细的统计逻辑
	return &HTTPClientStats{
		PoolSize:       1, // 简化实现
		ActiveClients:  0,
		IdleClients:    1,
		TotalRequests:  0,
		TotalErrors:    0,
		AverageLatency: 0,
		LastUsed:       time.Now(),
	}
}

// HTTPClientHealth HTTP客户端健康检查
type HTTPClientHealth struct {
	IsHealthy bool             `json:"is_healthy"`
	Message   string           `json:"message"`
	Stats     *HTTPClientStats `json:"stats"`
}

// GetHTTPClientHealth 获取HTTP客户端健康状态
func (gw *Gateway) GetHTTPClientHealth() *HTTPClientHealth {
	stats := gw.GetHTTPClientStats()

	isHealthy := true
	message := "HTTP客户端正常"

	// 简单的健康检查逻辑
	if stats.TotalErrors > 100 {
		isHealthy = false
		message = "HTTP客户端错误过多"
	}

	return &HTTPClientHealth{
		IsHealthy: isHealthy,
		Message:   message,
		Stats:     stats,
	}
}

// ResetHTTPClientPool 重置HTTP客户端池
func (gw *Gateway) ResetHTTPClientPool() {
	// 重新创建客户端池
	httpClientPool = &HTTPClientPool{
		pool: &sync.Pool{
			New: func() interface{} {
				config := getConfig()
				httpConfig := config.GetHTTPClientConfig()
				return &http.Client{
					Timeout: httpConfig.Timeout,
					Transport: &http.Transport{
						MaxIdleConns:          httpConfig.MaxIdleConns,
						MaxIdleConnsPerHost:   httpConfig.MaxIdleConnsPerHost,
						MaxConnsPerHost:       httpConfig.MaxConnsPerHost,
						IdleConnTimeout:       httpConfig.IdleConnTimeout,
						TLSHandshakeTimeout:   httpConfig.TLSHandshakeTimeout,
						ExpectContinueTimeout: httpConfig.ExpectContinueTimeout,
						DisableKeepAlives:     false,
						DisableCompression:    false,
					},
				}
			},
		},
	}
}
