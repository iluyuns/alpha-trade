package strategy

import (
	"context"
	"fmt"

	"github.com/iluyuns/alpha-trade/internal/domain/model"
)

// SimpleVolatility 简单波动策略
// 逻辑：价格波动超过阈值时触发交易
type SimpleVolatility struct {
	symbol           string
	threshold        model.Money // 波动阈值（百分比）
	lastPrice        model.Money
	positionQuantity model.Money
}

// NewSimpleVolatility 创建简单波动策略
func NewSimpleVolatility(symbol string, threshold model.Money) *SimpleVolatility {
	return &SimpleVolatility{
		symbol:           symbol,
		threshold:        threshold,
		lastPrice:        model.Zero(),
		positionQuantity: model.Zero(),
	}
}

// Name 策略名称
func (s *SimpleVolatility) Name() string {
	return "SimpleVolatility"
}

// OnCandle K线事件处理
func (s *SimpleVolatility) OnCandle(ctx context.Context, candle *model.Candle) (*TradeSignal, error) {
	// 忽略非目标标的
	if candle.Symbol != s.symbol {
		return nil, nil
	}

	currentPrice := candle.Close

	// 初始化：记录首个价格
	if s.lastPrice.IsZero() {
		s.lastPrice = currentPrice
		return nil, nil
	}

	// 计算价格变化百分比
	priceDiff := currentPrice.Sub(s.lastPrice)
	changePercent := priceDiff.Div(s.lastPrice).Abs()

	// 更新最后价格
	defer func() { s.lastPrice = currentPrice }()

	// 检查是否超过阈值
	if changePercent.LT(s.threshold) {
		return nil, nil
	}

	// 生成信号
	var signal Signal
	var reason string

	if priceDiff.IsPositive() {
		// 价格上涨超过阈值 -> 买入
		if s.positionQuantity.IsZero() {
			signal = SignalBuy
			reason = fmt.Sprintf("Price up %.2f%% (threshold: %.2f%%)",
				changePercent.Float64()*100, s.threshold.Float64()*100)
		} else {
			return nil, nil // 已持仓，不重复买入
		}
	} else {
		// 价格下跌超过阈值 -> 卖出
		if s.positionQuantity.IsPositive() {
			signal = SignalSell
			reason = fmt.Sprintf("Price down %.2f%% (threshold: %.2f%%)",
				changePercent.Float64()*100, s.threshold.Float64()*100)
		} else {
			return nil, nil // 无持仓，不卖出
		}
	}

	// 更新仓位（简化：固定数量）
	quantity := model.MustMoney("0.01")
	if signal == SignalBuy {
		s.positionQuantity = s.positionQuantity.Add(quantity)
	} else if signal == SignalSell {
		s.positionQuantity = s.positionQuantity.Sub(quantity)
	}

	return &TradeSignal{
		Signal:   signal,
		Symbol:   s.symbol,
		Price:    currentPrice,
		Quantity: quantity,
		Reason:   reason,
	}, nil
}

// OnTick Tick事件处理
func (s *SimpleVolatility) OnTick(ctx context.Context, tick *model.Tick) (*TradeSignal, error) {
	// 此策略不处理Tick
	return nil, nil
}
