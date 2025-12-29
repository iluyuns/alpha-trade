package telemetry

// Config 监控基础设施配置
type Config struct {
	JaegerEndpoint string
	PromPort       int
	Enabled        bool
	SampleRate     float64 // 采样率：0.0 - 1.0
}

// SetupTelemetry 组装并启动监控组件
func SetupTelemetry(cfg Config) {
	if !cfg.Enabled {
		return
	}
	// 调用 pkg/telemetry 完成初始化
}

