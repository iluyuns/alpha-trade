package metrics

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// MetricsHandler Prometheus metrics 处理器
func MetricsHandler() http.HandlerFunc {
	return promhttp.Handler().ServeHTTP
}
