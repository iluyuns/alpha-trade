package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// TradingMetrics 交易系统指标
type TradingMetrics struct {
	// 订单指标
	OrdersTotal     prometheus.Counter
	OrdersFilled    prometheus.Counter
	OrdersRejected  prometheus.Counter
	OrdersCancelled prometheus.Counter

	// 风控指标
	RiskChecksTotal      prometheus.Counter
	RiskChecksBlocked    prometheus.Counter
	RiskChecksAllowed    prometheus.Counter
	CircuitBreakerOpened prometheus.Counter

	// 盈亏指标
	PnLTotal   prometheus.Gauge
	PnLDaily   prometheus.Gauge
	PnLPercent prometheus.Gauge

	// 仓位指标
	TotalExposure   prometheus.Gauge
	PositionCount   prometheus.Gauge
	MaxPositionSize prometheus.Gauge

	// 系统指标
	GatewayLatency   prometheus.Histogram
	RiskCheckLatency prometheus.Histogram
	OrderLatency     prometheus.Histogram
}

var (
	// DefaultMetrics 默认指标实例
	DefaultMetrics *TradingMetrics
)

func init() {
	DefaultMetrics = NewTradingMetrics()
}

// NewTradingMetrics 创建交易系统指标
func NewTradingMetrics() *TradingMetrics {
	return &TradingMetrics{
		// 订单指标
		OrdersTotal: promauto.NewCounter(prometheus.CounterOpts{
			Name: "alpha_trade_orders_total",
			Help: "Total number of orders placed",
		}),
		OrdersFilled: promauto.NewCounter(prometheus.CounterOpts{
			Name: "alpha_trade_orders_filled_total",
			Help: "Total number of orders filled",
		}),
		OrdersRejected: promauto.NewCounter(prometheus.CounterOpts{
			Name: "alpha_trade_orders_rejected_total",
			Help: "Total number of orders rejected",
		}),
		OrdersCancelled: promauto.NewCounter(prometheus.CounterOpts{
			Name: "alpha_trade_orders_cancelled_total",
			Help: "Total number of orders cancelled",
		}),

		// 风控指标
		RiskChecksTotal: promauto.NewCounter(prometheus.CounterOpts{
			Name: "alpha_trade_risk_checks_total",
			Help: "Total number of risk checks performed",
		}),
		RiskChecksBlocked: promauto.NewCounter(prometheus.CounterOpts{
			Name: "alpha_trade_risk_checks_blocked_total",
			Help: "Total number of orders blocked by risk manager",
		}),
		RiskChecksAllowed: promauto.NewCounter(prometheus.CounterOpts{
			Name: "alpha_trade_risk_checks_allowed_total",
			Help: "Total number of orders allowed by risk manager",
		}),
		CircuitBreakerOpened: promauto.NewCounter(prometheus.CounterOpts{
			Name: "alpha_trade_circuit_breaker_opened_total",
			Help: "Total number of times circuit breaker was opened",
		}),

		// 盈亏指标
		PnLTotal: promauto.NewGauge(prometheus.GaugeOpts{
			Name: "alpha_trade_pnl_total",
			Help: "Total profit and loss (in USDT)",
		}),
		PnLDaily: promauto.NewGauge(prometheus.GaugeOpts{
			Name: "alpha_trade_pnl_daily",
			Help: "Daily profit and loss (in USDT)",
		}),
		PnLPercent: promauto.NewGauge(prometheus.GaugeOpts{
			Name: "alpha_trade_pnl_percent",
			Help: "Profit and loss percentage",
		}),

		// 仓位指标
		TotalExposure: promauto.NewGauge(prometheus.GaugeOpts{
			Name: "alpha_trade_total_exposure",
			Help: "Total exposure (in USDT)",
		}),
		PositionCount: promauto.NewGauge(prometheus.GaugeOpts{
			Name: "alpha_trade_position_count",
			Help: "Number of open positions",
		}),
		MaxPositionSize: promauto.NewGauge(prometheus.GaugeOpts{
			Name: "alpha_trade_max_position_size",
			Help: "Maximum position size (in USDT)",
		}),

		// 系统指标
		GatewayLatency: promauto.NewHistogram(prometheus.HistogramOpts{
			Name:    "alpha_trade_gateway_latency_seconds",
			Help:    "Gateway operation latency in seconds",
			Buckets: prometheus.DefBuckets,
		}),
		RiskCheckLatency: promauto.NewHistogram(prometheus.HistogramOpts{
			Name:    "alpha_trade_risk_check_latency_seconds",
			Help:    "Risk check latency in seconds",
			Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1.0},
		}),
		OrderLatency: promauto.NewHistogram(prometheus.HistogramOpts{
			Name:    "alpha_trade_order_latency_seconds",
			Help:    "Order processing latency in seconds",
			Buckets: prometheus.DefBuckets,
		}),
	}
}
