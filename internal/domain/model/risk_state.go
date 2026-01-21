package model

import "time"

// RiskState 风控状态（持久化到 Redis/DB）
type RiskState struct {
	// 账户级别
	AccountID string
	Symbol    string // 标的（空字符串表示账户全局状态）

	// 初始/当前净值
	InitialEquity Money // 初始净值（用于计算总回报）
	CurrentEquity Money // 当前净值
	PeakEquity    Money // 历史峰值净值

	// 当日盈亏统计
	DailyPnL        Money     // 当日累计盈亏
	DailyTradeCount int       // 当日交易次数
	DailyResetTime  time.Time // 每日重置时间（UTC 0点）
	LastResetDate   string    // 最后重置日期（格式：2006-01-02）

	// 连续亏损跟踪
	ConsecutiveLosses int       // 连续亏损次数
	LastLossTime      time.Time // 最后一次亏损时间

	// 最大回撤（MaxDrawDown）
	MDD        Money // 当前最大回撤金额
	MDDPercent Money // 当前最大回撤百分比

	// 仓位统计
	TotalExposure Money            // 总敞口（所有持仓名义价值之和）
	PositionMap   map[string]Money // 各标的持仓（symbol -> notional value）

	// 熔断状态
	CircuitBreakerOpen  bool  // 熔断器是否打开
	CircuitBreakerUntil int64 // 熔断解除时间（Unix时间戳）

	// 快照时间
	UpdatedAt time.Time
}

// NewRiskState 创建初始风控状态
func NewRiskState(accountID string, initialEquity Money) *RiskState {
	now := time.Now()
	return &RiskState{
		AccountID:      accountID,
		Symbol:         "", // 默认为账户全局状态
		InitialEquity:  initialEquity,
		CurrentEquity:  initialEquity,
		PeakEquity:     initialEquity,
		DailyPnL:       Zero(),
		MDD:            Zero(),
		MDDPercent:     Zero(),
		PositionMap:    make(map[string]Money),
		TotalExposure:  Zero(),
		UpdatedAt:      now,
		DailyResetTime: getNextDayUTC(),
		LastResetDate:  now.Format("2006-01-02"),
	}
}

// UpdateEquity 更新净值并计算MDD
func (rs *RiskState) UpdateEquity(newEquity Money) {
	rs.CurrentEquity = newEquity

	// 更新峰值
	if newEquity.GT(rs.PeakEquity) {
		rs.PeakEquity = newEquity
	}

	// 计算当前回撤
	if rs.PeakEquity.IsPositive() {
		drawdown := rs.PeakEquity.Sub(newEquity)
		rs.MDD = drawdown
		rs.MDDPercent = drawdown.Div(rs.PeakEquity)
	}

	rs.UpdatedAt = time.Now()
}

// RecordLoss 记录亏损（连续亏损计数）
func (rs *RiskState) RecordLoss() {
	rs.ConsecutiveLosses++
	rs.LastLossTime = time.Now()
	rs.UpdatedAt = time.Now()
}

// ResetConsecutiveLosses 重置连续亏损（盈利时调用）
func (rs *RiskState) ResetConsecutiveLosses() {
	rs.ConsecutiveLosses = 0
	rs.UpdatedAt = time.Now()
}

// OpenCircuitBreaker 打开熔断器
func (rs *RiskState) OpenCircuitBreaker(duration time.Duration) {
	rs.CircuitBreakerOpen = true
	rs.CircuitBreakerUntil = time.Now().Add(duration).Unix()
	rs.UpdatedAt = time.Now()
}

// CloseCircuitBreaker 关闭熔断器
func (rs *RiskState) CloseCircuitBreaker() {
	rs.CircuitBreakerOpen = false
	rs.CircuitBreakerUntil = 0
	rs.UpdatedAt = time.Now()
}

// ShouldResetDaily 是否需要每日重置
func (rs *RiskState) ShouldResetDaily(now time.Time) bool {
	return now.After(rs.DailyResetTime)
}

// ResetDaily 每日重置
func (rs *RiskState) ResetDaily() {
	now := time.Now()
	rs.DailyPnL = Zero()
	rs.DailyTradeCount = 0
	rs.DailyResetTime = getNextDayUTC()
	rs.LastResetDate = now.Format("2006-01-02")
	rs.UpdatedAt = now
}

// getNextDayUTC 获取下一个UTC午夜时间
func getNextDayUTC() time.Time {
	now := time.Now().UTC()
	tomorrow := now.Add(24 * time.Hour)
	return time.Date(tomorrow.Year(), tomorrow.Month(), tomorrow.Day(), 0, 0, 0, 0, time.UTC)
}
