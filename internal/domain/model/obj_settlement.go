package model

import (
	"time"

	"github.com/shopspring/decimal"
)

// Settlement 统一结算单模型 (支持 Spot 和 Future)
type Settlement struct {
	ID         int64
	TradeID    string // 逻辑交易 ID (FIFO 匹配的一对多或多对多平仓记录)
	Symbol     string
	MarketType string // "SPOT" 或 "FUTURE"
	Side       string // "LONG" 或 "SHORT"

	// 核心财务
	RealizedPnL decimal.Decimal // 已实现盈亏 (净值)
	Commission  decimal.Decimal // 累计手续费支出
	FundingFee  decimal.Decimal // 累计资金费支出/收入 (Spot 为 0)

	// 绩效指标
	EntryPrice decimal.Decimal
	ExitPrice  decimal.Decimal
	Quantity   decimal.Decimal
	ROI        decimal.Decimal

	// 时间归归因
	OpenedAt        time.Time
	ClosedAt        time.Time
	DurationSeconds int64

	// 扩展元数据
	Metadata map[string]interface{}
}

// NewSettlementFromFuture 创建合约结算单
func NewSettlementFromFuture(symbol string, pnl, commission, funding decimal.Decimal) *Settlement {
	return &Settlement{
		Symbol:     symbol,
		MarketType: "FUTURE",
		RealizedPnL: pnl,
		Commission:  commission,
		FundingFee:  funding,
	}
}

// NewSettlementFromSpot 创建现货结算单
func NewSettlementFromSpot(symbol string, pnl, commission decimal.Decimal) *Settlement {
	return &Settlement{
		Symbol:     symbol,
		MarketType: "SPOT",
		RealizedPnL: pnl,
		Commission:  commission,
		FundingFee:  decimal.Zero,
	}
}

