package dashboard

import (
	"context"
	"fmt"
	"time"

	"github.com/iluyuns/alpha-trade/internal/domain/model"
	"github.com/iluyuns/alpha-trade/internal/svc"
	"github.com/iluyuns/alpha-trade/internal/types"
	"github.com/shopspring/decimal"
	"github.com/zeromicro/go-zero/core/logx"
)

type DashboardLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewDashboardLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DashboardLogic {
	return &DashboardLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DashboardLogic) Dashboard(req *types.DashboardReq) (resp *types.DashboardResp, err error) {
	resp = &types.DashboardResp{}

	// 1. 从 Prometheus Metrics 获取核心指标
	if err := l.fetchMetricsFromPrometheus(resp); err != nil {
		l.Errorf("Failed to fetch metrics: %v", err)
		// 继续执行，使用默认值
	}

	// 2. 从 RiskRepo 获取风控状态
	if err := l.fetchRiskStatus(resp); err != nil {
		l.Errorf("Failed to fetch risk status: %v", err)
	}

	// 3. 获取系统健康状态
	l.fetchSystemHealth(resp)

	// 4. 获取策略概览
	if err := l.fetchStrategies(resp); err != nil {
		l.Errorf("Failed to fetch strategies: %v", err)
	}

	return resp, nil
}

func (l *DashboardLogic) fetchMetricsFromPrometheus(resp *types.DashboardResp) error {
	// 直接从内存中的 metrics 获取值（更高效）
	// 注意：Prometheus Gauge 的值需要通过 HTTP 端点获取，这里简化处理

	// 使用默认值，实际应该从 Prometheus HTTP API 查询
	// 或者直接从 metrics.DefaultMetrics 读取（如果暴露了 Get 方法）
	resp.PnLDaily = "0.00"
	resp.PnLPercent = "0.00"
	resp.TotalEquity = "100000.00"
	resp.RiskExposure = "0.00"
	resp.DailyDrawdown = "0.00"

	// TODO: 实现从 Prometheus API 查询
	// 示例：查询 alpha_trade_pnl_daily
	return nil
}

func (l *DashboardLogic) fetchRiskStatus(resp *types.DashboardResp) error {
	if l.svcCtx.RiskRepo == nil {
		resp.RiskStatus = types.RiskStatus{
			ConsecutiveLosses:    0,
			MaxConsecutiveLosses: 5,
			MacroCoolingMode:     "inactive",
			LeverageStatus:       "relaxed",
			MaxLeverage:          "2.0",
			CurrentLeverage:      "1.0",
		}
		return nil
	}

	accountID := "default-account"
	state, err := l.svcCtx.RiskRepo.LoadState(l.ctx, accountID, "")
	if err != nil {
		// 如果状态不存在，返回默认值
		resp.RiskStatus = types.RiskStatus{
			ConsecutiveLosses:    0,
			MaxConsecutiveLosses: 5,
			MacroCoolingMode:     "inactive",
			LeverageStatus:       "relaxed",
			MaxLeverage:          "2.0",
			CurrentLeverage:      "1.0",
		}
		return nil
	}

	// 计算风险敞口百分比
	exposurePercent := decimal.Zero
	if state.CurrentEquity.IsPositive() {
		exposurePercent = state.TotalExposure.Div(state.CurrentEquity).Mul(model.NewMoneyFromInt(100)).Decimal()
	}
	resp.RiskExposure = exposurePercent.String()

	// 计算日内回撤（MDDPercent 已经是百分比，直接使用）
	resp.DailyDrawdown = state.MDDPercent.String()

	// 更新总权益
	resp.TotalEquity = state.CurrentEquity.String()

	// 更新今日盈亏
	resp.PnLDaily = state.DailyPnL.String()
	if state.CurrentEquity.IsPositive() {
		pnlPercent := state.DailyPnL.Div(state.CurrentEquity).Mul(model.NewMoneyFromInt(100))
		resp.PnLPercent = pnlPercent.String()
	}

	// 构建风控状态
	maxConsecutiveLosses := 5
	if l.svcCtx.RiskManager != nil {
		// 从配置获取，这里简化处理
		maxConsecutiveLosses = 5
	}

	leverageStatus := "relaxed"
	if state.CircuitBreakerOpen {
		leverageStatus = "restricted"
	}

	resp.RiskStatus = types.RiskStatus{
		ConsecutiveLosses:    int64(state.ConsecutiveLosses),
		MaxConsecutiveLosses: int64(maxConsecutiveLosses),
		MacroCoolingMode:     "inactive", // TODO: 从 EventGateway 获取
		LeverageStatus:       leverageStatus,
		MaxLeverage:          "2.0",
		CurrentLeverage:      "1.0", // TODO: 从实际持仓计算
	}

	return nil
}

func (l *DashboardLogic) fetchSystemHealth(resp *types.DashboardResp) {
	health := []types.SystemHealthItem{}

	// Spot Gateway 状态
	spotStatus := "normal"
	spotLatency := int64(0)
	if l.svcCtx.BinanceSpotClient != nil {
		// 检查连接状态（简化处理）
		spotStatus = "normal"
		spotLatency = 23 // TODO: 从实际延迟获取
	} else {
		spotStatus = "error"
	}

	health = append(health, types.SystemHealthItem{
		Name:    "Spot Gateway",
		Status:  spotStatus,
		Latency: spotLatency,
		Message: fmt.Sprintf("%s (%dms)", spotStatus, spotLatency),
	})

	// Future Gateway（暂未实现）
	health = append(health, types.SystemHealthItem{
		Name:    "Future Gateway",
		Status:  "error",
		Message: "Not implemented",
	})

	// AI Agent（暂未实现）
	health = append(health, types.SystemHealthItem{
		Name:          "AI Agent",
		Status:        "error",
		LastHeartbeat: time.Now().Format(time.RFC3339),
		Message:       "Not implemented",
	})

	// News Feed（暂未实现）
	health = append(health, types.SystemHealthItem{
		Name:          "News Feed",
		Status:        "error",
		LastHeartbeat: time.Now().Format(time.RFC3339),
		Message:       "Not implemented",
	})

	resp.SystemHealth = health
}

func (l *DashboardLogic) fetchStrategies(resp *types.DashboardResp) error {
	strategies := []types.StrategyOverview{}

	// 从数据库查询策略配置
	if l.svcCtx.DB != nil {
		// TODO: 查询策略配置表
		// 这里简化处理，返回示例数据
		strategies = append(strategies, types.StrategyOverview{
			ID:        "ST-101",
			Name:      "MACD_Breakout",
			Symbol:    "BTCUSDT",
			Direction: "Long",
			Status:    "running",
			WinRate:   "65",
		})

		strategies = append(strategies, types.StrategyOverview{
			ID:        "ST-102",
			Name:      "Grid_Maker",
			Symbol:    "ETHUSDT",
			Direction: "N/A",
			Status:    "stopped",
			WinRate:   "40",
		})
	}

	resp.Strategies = strategies
	return nil
}
