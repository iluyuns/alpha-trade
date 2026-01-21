package port

import (
	"context"

	"github.com/iluyuns/alpha-trade/internal/domain/model"
)

// RiskRepo 风控状态持久化接口
// 实现要求：
// 1. 幂等写入（相同状态多次写入结果一致）
// 2. 原子更新（状态变更全部成功或全部失败）
// 3. 支持回测（内存实现）与实盘（Redis/DB实现）
type RiskRepo interface {
	// LoadState 加载风控状态
	// accountID: 账户ID（支持多账户）
	// symbol: 标的（空字符串表示全局状态）
	LoadState(ctx context.Context, accountID, symbol string) (*model.RiskState, error)

	// SaveState 保存风控状态（幂等）
	// 覆盖写入，确保状态一致性
	SaveState(ctx context.Context, state *model.RiskState) error

	// UpdateEquity 原子更新净值
	// 同时重新计算 MDD
	UpdateEquity(ctx context.Context, accountID string, newEquity model.Money) error

	// RecordTrade 记录交易（更新当日统计）
	// pnl: 本次交易盈亏
	RecordTrade(ctx context.Context, accountID string, pnl model.Money) error

	// OpenCircuitBreaker 打开熔断器
	// duration: 熔断持续时间
	OpenCircuitBreaker(ctx context.Context, accountID string, duration int64) error

	// CloseCircuitBreaker 关闭熔断器
	CloseCircuitBreaker(ctx context.Context, accountID string) error

	// IsCircuitBreakerOpen 检查熔断器状态
	IsCircuitBreakerOpen(ctx context.Context, accountID string) (bool, error)
}
