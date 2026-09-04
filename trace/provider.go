package trace

import (
	"context"
	"fmt"
	"strings"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
)

// Shutdown 关闭 TracerProvider 并刷新剩余 span。
type Shutdown func(context.Context) error

// Init 安装全局 TracerProvider 与 W3C Propagator。
// enabled=false 时不改动全局 Provider，返回空 Shutdown。
func Init(ctx context.Context, cfg Config) (Shutdown, error) {
	if !cfg.Enabled {
		return noopShutdown, nil
	}
	if strings.TrimSpace(cfg.ServiceName) == "" {
		return nil, fmt.Errorf("trace: service name is required when enabled")
	}

	kv := []attribute.KeyValue{semconv.ServiceName(cfg.ServiceName)}
	if v := strings.TrimSpace(cfg.ServiceVersion); v != "" {
		kv = append(kv, semconv.ServiceVersion(v))
	}
	if v := strings.TrimSpace(cfg.Environment); v != "" {
		kv = append(kv, semconv.DeploymentEnvironment(v))
	}
	if v := strings.TrimSpace(cfg.System); v != "" {
		kv = append(kv, attribute.String("system", v))
	}
	res, err := resource.New(ctx, resource.WithAttributes(kv...))
	if err != nil {
		return nil, fmt.Errorf("trace: resource: %w", err)
	}

	opts := []sdktrace.TracerProviderOption{sdktrace.WithResource(res)}
	if s := cfg.sampleRate(); s < 1 {
		opts = append(opts, sdktrace.WithSampler(sdktrace.TraceIDRatioBased(s)))
	}

	switch strings.ToLower(strings.TrimSpace(cfg.Exporter)) {
	case "", "none":
		// 层 1：进程内 span 树，不上报。
	case "otlp":
		endpoint := cleanEndpoint(cfg.Endpoint)
		if endpoint == "" {
			return nil, fmt.Errorf("trace: exporter=otlp requires endpoint")
		}
		exp, err := otlptracegrpc.New(ctx, otlptracegrpc.WithEndpoint(endpoint), otlptracegrpc.WithInsecure())
		if err != nil {
			return nil, fmt.Errorf("trace: otlp exporter: %w", err)
		}
		opts = append(opts, sdktrace.WithBatcher(exp))
	default:
		return nil, fmt.Errorf("trace: unsupported exporter %q", cfg.Exporter)
	}

	tp := sdktrace.NewTracerProvider(opts...)
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))
	return tp.Shutdown, nil
}

func noopShutdown(context.Context) error { return nil }

func cleanEndpoint(ep string) string {
	ep = strings.TrimSpace(ep)
	ep = strings.TrimPrefix(ep, "http://")
	ep = strings.TrimPrefix(ep, "https://")
	return strings.TrimSuffix(ep, "/")
}
