package model

import (
	"github.com/shopspring/decimal"
)

type FeeType int

const (
	FeeTypeMaker FeeType = iota
	FeeTypeTaker
)

// SymbolConfig 定义交易对的静态属性
type SymbolConfig struct {
	Name       string // e.g., "BTCUSDT"
	MarketType string // "SPOT" 或 "FUTURE"
	BaseAsset  string // "BTC"
	QuoteAsset string // "USDT"

	// 风险配置 (Risk Protocol Enforcement)
	MarginMode  string          // 强制 "ISOLATED" (仅 FUTURE)
	MaxLeverage decimal.Decimal // 强制 <= 2.0 (仅 FUTURE)

	// 费率配置 (优先使用 ConfiguredFee 进行回测/保守估算)
	UseCustomFee   bool            // 是否强制使用手动费率
	CustomMakerFee decimal.Decimal // 手动配置 Maker 费率 (e.g. 0.0002)
	CustomTakerFee decimal.Decimal // 手动配置 Taker 费率 (e.g. 0.0005)

	// 精度配置
	PricePrecision int32
	QtyPrecision   int32
	MinQty         decimal.Decimal
}

// GetEffectiveFee 返回有效的手续费率
func (s *SymbolConfig) GetEffectiveFee(ft FeeType, realMaker, realTaker decimal.Decimal) decimal.Decimal {
	if s.UseCustomFee {
		if ft == FeeTypeMaker {
			return s.CustomMakerFee
		}
		return s.CustomTakerFee
	}
	if ft == FeeTypeMaker {
		return realMaker
	}
	return realTaker
}

