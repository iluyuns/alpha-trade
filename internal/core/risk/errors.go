package risk

import "errors"

var (
	// ErrRiskLimitExceeded 风控限额超限
	ErrRiskLimitExceeded = errors.New("risk: limit exceeded")

	// ErrCircuitBreakerOpen 熔断器打开
	ErrCircuitBreakerOpen = errors.New("risk: circuit breaker open")

	// ErrInvalidOrder 订单参数无效
	ErrInvalidOrder = errors.New("risk: invalid order parameters")

	// ErrPositionLimitExceeded 仓位限制超限
	ErrPositionLimitExceeded = errors.New("risk: position limit exceeded")

	// ErrInsufficientMargin 保证金不足
	ErrInsufficientMargin = errors.New("risk: insufficient margin")

	// ErrFatFinger 胖手指错误（价格/数量异常）
	ErrFatFinger = errors.New("risk: fat finger detected")

	// ErrMacroEventCooldown 宏观事件冷却期
	ErrMacroEventCooldown = errors.New("risk: macro event cooldown active")

	// ErrMaxDrawdownExceeded 最大回撤超限
	ErrMaxDrawdownExceeded = errors.New("risk: max drawdown exceeded")
)
