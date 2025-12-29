package telemetry

import (
	"context"
	"time"
)

// TODO: 待接入 go.opentelemetry.io/otel 相关依赖后实现
// InitTracer 初始化全局 Tracer
func InitTracer(serviceName string, endpoint string) (func(context.Context) error, error) {
	// 1. 创建 Jaeger Exporter
	// 2. 注册 TracerProvider
	// 3. 设置全局 Propagator (用于跨进程传递 TraceID)
	return func(ctx context.Context) error { return nil }, nil
}

// InitMeter 初始化全局 Metrics 采集器
func InitMeter(serviceName string) error {
	// 1. 创建 Prometheus Exporter
	// 2. 注册 MeterProvider
	return nil
}

