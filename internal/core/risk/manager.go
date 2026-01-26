package risk

import (
	"context"
	"sync"
	"time"

	"github.com/iluyuns/alpha-trade/internal/domain/model"
	"github.com/iluyuns/alpha-trade/internal/domain/port"
	"github.com/iluyuns/alpha-trade/internal/pkg/metrics"
)

// OrderContext 订单上下文（风控检查输入）
type OrderContext struct {
	// 基础信息
	ClientOrderID string
	Symbol        string
	MarketType    model.MarketType
	Side          model.OrderSide
	Type          model.OrderType

	// 价格与数量
	Price    model.Money
	Quantity model.Money

	// 合约专属
	Leverage   int
	ReduceOnly bool

	// 风控参数
	ProtectPrice model.Money // 保护价
	CurrentPrice model.Money // 当前市价（用于计算名义价值）
	AccountID    string      // 账户ID
}

// RiskConfig 风控配置快照（不可变）
type RiskConfig struct {
	// 熔断器配置
	MaxConsecutiveLosses int     // 最大连续亏损次数
	MaxDailyDrawdown     float64 // 最大日内回撤百分比
	MaxTotalMDD          float64 // 最大总回撤百分比

	// 仓位限制
	MaxSinglePositionPercent float64 // 单标的最大仓位占比
	MaxTotalExposurePercent  float64 // 总敞口最大占比
	MinCashReservePercent    float64 // 最小现金储备占比

	// 合约限制
	MaxLeverage         int     // 最大杠杆倍数
	ForceLeverageOne    bool    // 大额单强制1x
	LargeOrderThreshold float64 // 大额单阈值（占账户比例）

	// Fat Finger 检测
	MaxPriceDeviation float64 // 最大价格偏离（百分比）
	MaxOrderNotional  float64 // 单笔最大名义价值（USD）
}

// Manager 风控管理器
type Manager struct {
	mu     sync.RWMutex
	repo   port.RiskRepo
	config RiskConfig

	// 内存缓存（减少 IO）
	stateCache map[string]*model.RiskState
}

// NewManager 创建风控管理器
func NewManager(repo port.RiskRepo, config RiskConfig) *Manager {
	return &Manager{
		repo:       repo,
		config:     config,
		stateCache: make(map[string]*model.RiskState),
	}
}

// CheckPreTrade 交易前风控检查（核心入口）
// 按规则顺序短路评估：CircuitBreaker -> PositionLimit -> FatFinger
func (m *Manager) CheckPreTrade(ctx context.Context, req *OrderContext) (DecisionDetail, error) {
	startTime := time.Now()
	metrics.DefaultMetrics.RiskChecksTotal.Inc()

	// 1. 加载风控状态
	state, err := m.loadState(ctx, req.AccountID, "")
	if err != nil {
		metrics.DefaultMetrics.RiskChecksBlocked.Inc()
		return NewBlock("failed to load risk state", "internal"), err
	}

	// 2. 每日重置检查
	if state.ShouldResetDaily(time.Now()) {
		state.ResetDaily()
		_ = m.repo.SaveState(ctx, state)
	}

	// 3. 规则链检查（短路评估）
	rules := []RuleFunc{
		m.checkCircuitBreaker,
		m.checkPositionLimit,
		// m.checkFatFinger, // TODO: implement in rule_fat_finger.go
	}

	for _, rule := range rules {
		decision := rule(ctx, req, state)
		if !decision.IsAllowed() {
			// 记录延迟
			metrics.DefaultMetrics.RiskCheckLatency.Observe(time.Since(startTime).Seconds())
			metrics.DefaultMetrics.RiskChecksBlocked.Inc()
			if decision.IsBlocked() && decision.TriggeredRule == "circuit_breaker" {
				metrics.DefaultMetrics.CircuitBreakerOpened.Inc()
			}
			return decision, nil
		}
	}

	// 记录延迟
	metrics.DefaultMetrics.RiskCheckLatency.Observe(time.Since(startTime).Seconds())
	metrics.DefaultMetrics.RiskChecksAllowed.Inc()

	return NewAllow(), nil
}

// RuleFunc 风控规则函数签名
type RuleFunc func(ctx context.Context, req *OrderContext, state *model.RiskState) DecisionDetail

// loadState 加载风控状态（带缓存）
func (m *Manager) loadState(ctx context.Context, accountID, symbol string) (*model.RiskState, error) {
	m.mu.RLock()
	cacheKey := accountID + ":" + symbol
	cached, exists := m.stateCache[cacheKey]
	m.mu.RUnlock()

	if exists {
		return cached, nil
	}

	// 从持久化加载
	state, err := m.repo.LoadState(ctx, accountID, symbol)
	if err != nil {
		return nil, err
	}

	// 写入缓存
	m.mu.Lock()
	m.stateCache[cacheKey] = state
	m.mu.Unlock()

	return state, nil
}

// InvalidateCache 清除缓存（状态变更后调用）
func (m *Manager) InvalidateCache(accountID, symbol string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.stateCache, accountID+":"+symbol)
}

// checkFatFinger Fat Finger 检测（待实现）
func (m *Manager) checkFatFinger(ctx context.Context, req *OrderContext, state *model.RiskState) DecisionDetail {
	// TODO: implement in rule_fat_finger.go
	return NewAllow()
}
