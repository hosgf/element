package trace

// Config OTel 初始化配置（层 1 传播 + 层 2 可选上报，见分布式链路追踪标准 §4.3）。
type Config struct {
	// Enabled 层 1：启用 TracerProvider 与 W3C traceparent 传播。
	Enabled bool
	// Exporter 层 2：none | otlp；none 时 span End 后丢弃，不依赖 Collector。
	Exporter string
	// Endpoint exporter=otlp 时必填，如 otel-collector:4317。
	Endpoint string
	// ServiceName 进程 service.name。
	ServiceName string
	// ServiceVersion 可选，写入 resource。
	ServiceVersion string
	// SampleRate 采样率，主要影响层 2 导出量；默认 1.0。
	SampleRate float64
}

func (c Config) sampleRate() float64 {
	if c.SampleRate <= 0 || c.SampleRate > 1 {
		return 1.0
	}
	return c.SampleRate
}
