package risk

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/iluyuns/alpha-trade/internal/domain/model"
	"github.com/iluyuns/alpha-trade/internal/domain/port"
)

// MemoryRiskRepo 内存风控状态仓储（回测/测试用）
type MemoryRiskRepo struct {
	mu     sync.RWMutex
	states map[string]*model.RiskState // key: accountID:symbol
}

// NewMemoryRiskRepo 创建内存仓储
func NewMemoryRiskRepo() port.RiskRepo {
	return &MemoryRiskRepo{
		states: make(map[string]*model.RiskState),
	}
}

// LoadState 加载风控状态
func (r *MemoryRiskRepo) LoadState(ctx context.Context, accountID, symbol string) (*model.RiskState, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	key := makeKey(accountID, symbol)
	state, exists := r.states[key]
	if !exists {
		// 不存在则返回默认状态
		return model.NewRiskState(accountID, model.MustMoney("10000")), nil
	}

	// 返回副本，避免外部修改
	return copyState(state), nil
}

// SaveState 保存风控状态（幂等）
func (r *MemoryRiskRepo) SaveState(ctx context.Context, state *model.RiskState) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	key := makeKey(state.AccountID, "")
	r.states[key] = copyState(state)
	return nil
}

// UpdateEquity 原子更新净值
func (r *MemoryRiskRepo) UpdateEquity(ctx context.Context, accountID string, newEquity model.Money) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	key := makeKey(accountID, "")
	state, exists := r.states[key]
	if !exists {
		state = model.NewRiskState(accountID, newEquity)
	}

	state.UpdateEquity(newEquity)
	r.states[key] = state
	return nil
}

// RecordTrade 记录交易（更新当日统计）
func (r *MemoryRiskRepo) RecordTrade(ctx context.Context, accountID string, pnl model.Money) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	key := makeKey(accountID, "")
	state, exists := r.states[key]
	if !exists {
		return fmt.Errorf("account %s not found", accountID)
	}

	// 更新当日盈亏
	state.DailyPnL = state.DailyPnL.Add(pnl)
	state.DailyTradeCount++

	// 更新连续亏损
	if pnl.IsNegative() {
		state.RecordLoss()
	} else if pnl.IsPositive() {
		state.ResetConsecutiveLosses()
	}

	state.UpdatedAt = time.Now()
	r.states[key] = state
	return nil
}

// OpenCircuitBreaker 打开熔断器
func (r *MemoryRiskRepo) OpenCircuitBreaker(ctx context.Context, accountID string, durationSec int64) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	key := makeKey(accountID, "")
	state, exists := r.states[key]
	if !exists {
		return fmt.Errorf("account %s not found", accountID)
	}

	state.OpenCircuitBreaker(time.Duration(durationSec) * time.Second)
	r.states[key] = state
	return nil
}

// CloseCircuitBreaker 关闭熔断器
func (r *MemoryRiskRepo) CloseCircuitBreaker(ctx context.Context, accountID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	key := makeKey(accountID, "")
	state, exists := r.states[key]
	if !exists {
		return fmt.Errorf("account %s not found", accountID)
	}

	state.CloseCircuitBreaker()
	r.states[key] = state
	return nil
}

// IsCircuitBreakerOpen 检查熔断器状态
func (r *MemoryRiskRepo) IsCircuitBreakerOpen(ctx context.Context, accountID string) (bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	key := makeKey(accountID, "")
	state, exists := r.states[key]
	if !exists {
		return false, nil
	}

	if state.CircuitBreakerOpen && time.Now().Unix() < state.CircuitBreakerUntil {
		return true, nil
	}

	return false, nil
}

// makeKey 生成存储键
func makeKey(accountID, symbol string) string {
	if symbol == "" {
		return accountID
	}
	return accountID + ":" + symbol
}

// copyState 深拷贝状态（避免并发修改）
func copyState(state *model.RiskState) *model.RiskState {
	copied := &model.RiskState{
		AccountID:           state.AccountID,
		Symbol:              state.Symbol,
		InitialEquity:       state.InitialEquity,
		CurrentEquity:       state.CurrentEquity,
		PeakEquity:          state.PeakEquity,
		DailyPnL:            state.DailyPnL,
		DailyTradeCount:     state.DailyTradeCount,
		DailyResetTime:      state.DailyResetTime,
		LastResetDate:       state.LastResetDate,
		ConsecutiveLosses:   state.ConsecutiveLosses,
		LastLossTime:        state.LastLossTime,
		MDD:                 state.MDD,
		MDDPercent:          state.MDDPercent,
		TotalExposure:       state.TotalExposure,
		CircuitBreakerOpen:  state.CircuitBreakerOpen,
		CircuitBreakerUntil: state.CircuitBreakerUntil,
		UpdatedAt:           state.UpdatedAt,
		PositionMap:         make(map[string]model.Money),
	}

	// 拷贝 map
	for k, v := range state.PositionMap {
		copied.PositionMap[k] = v
	}

	return copied
}
