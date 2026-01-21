package risk

import (
	"context"
	"fmt"
	"time"

	"github.com/iluyuns/alpha-trade/internal/domain/model"
)

// CheckCircuitBreaker 熔断器规则
// 触发条件：
// 1. 连续亏损次数 >= MaxConsecutiveLosses
// 2. 当日MDD >= MaxDailyDrawdown
// 3. 总MDD >= MaxTotalMDD
// 4. 熔断器已打开且未过期
func (m *Manager) CheckCircuitBreaker(ctx context.Context, req *OrderContext, state *model.RiskState) DecisionDetail {
	// 1. 检查熔断器是否已打开
	if state.CircuitBreakerOpen {
		if time.Now().Unix() < state.CircuitBreakerUntil {
			untilTime := time.Unix(state.CircuitBreakerUntil, 0)
			return NewBlock(
				fmt.Sprintf("circuit breaker active until %s", untilTime.Format(time.RFC3339)),
				"CircuitBreaker",
			)
		}
		// 熔断器已过期，自动关闭
		state.CloseCircuitBreaker()
		_ = m.repo.SaveState(ctx, state)
	}

	// 2. 检查连续亏损次数
	if m.config.MaxConsecutiveLosses > 0 && state.ConsecutiveLosses >= m.config.MaxConsecutiveLosses {
		// 打开熔断器（1小时冷却）
		state.OpenCircuitBreaker(1 * time.Hour)
		_ = m.repo.SaveState(ctx, state)

		return NewBlock(
			fmt.Sprintf("consecutive losses (%d) >= max allowed (%d)",
				state.ConsecutiveLosses, m.config.MaxConsecutiveLosses),
			"CircuitBreaker:ConsecutiveLosses",
		)
	}

	// 3. 检查当日回撤
	if m.config.MaxDailyDrawdown > 0 && state.CurrentEquity.IsPositive() {
		dailyPnLPercent := state.DailyPnL.Div(state.CurrentEquity).Float64()
		if dailyPnLPercent < -m.config.MaxDailyDrawdown {
			state.OpenCircuitBreaker(24 * time.Hour) // 次日重置
			_ = m.repo.SaveState(ctx, state)

			return NewBlock(
				fmt.Sprintf("daily drawdown (%.2f%%) >= max allowed (%.2f%%)",
					dailyPnLPercent*100, m.config.MaxDailyDrawdown*100),
				"CircuitBreaker:DailyDrawdown",
			)
		}
	}

	// 4. 检查总回撤
	if m.config.MaxTotalMDD > 0 {
		totalMDDPercent := state.MDDPercent.Float64()
		if totalMDDPercent >= m.config.MaxTotalMDD {
			state.OpenCircuitBreaker(7 * 24 * time.Hour) // 7天冷却
			_ = m.repo.SaveState(ctx, state)

			return NewBlock(
				fmt.Sprintf("total MDD (%.2f%%) >= max allowed (%.2f%%)",
					totalMDDPercent*100, m.config.MaxTotalMDD*100),
				"CircuitBreaker:TotalMDD",
			)
		}
	}

	return NewAllow()
}

// checkCircuitBreaker 内部调用（manager.go 中的短路链）
func (m *Manager) checkCircuitBreaker(ctx context.Context, req *OrderContext, state *model.RiskState) DecisionDetail {
	return m.CheckCircuitBreaker(ctx, req, state)
}
