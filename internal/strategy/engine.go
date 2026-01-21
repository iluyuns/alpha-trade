package strategy

import (
	"context"

	"github.com/google/uuid"
	"github.com/iluyuns/alpha-trade/internal/domain/model"
	"github.com/iluyuns/alpha-trade/internal/domain/port"
)

// Signal 交易信号
type Signal int

const (
	SignalNone Signal = iota
	SignalBuy
	SignalSell
	SignalHold
)

func (s Signal) String() string {
	switch s {
	case SignalBuy:
		return "BUY"
	case SignalSell:
		return "SELL"
	case SignalHold:
		return "HOLD"
	default:
		return "NONE"
	}
}

// TradeSignal 具体交易信号
type TradeSignal struct {
	Signal   Signal
	Symbol   string
	Price    model.Money
	Quantity model.Money
	Reason   string
}

// Strategy 策略接口
type Strategy interface {
	// Name 策略名称
	Name() string

	// OnCandle K线事件处理
	OnCandle(ctx context.Context, candle *model.Candle) (*TradeSignal, error)

	// OnTick Tick事件处理（可选）
	OnTick(ctx context.Context, tick *model.Tick) (*TradeSignal, error)
}

// Engine 策略引擎
type Engine struct {
	strategy      Strategy
	spotGateway   port.SpotGateway
	futureGateway port.FutureGateway
	accountID     string
}

// NewEngine 创建策略引擎
func NewEngine(strategy Strategy, spotGateway port.SpotGateway, accountID string) *Engine {
	return &Engine{
		strategy:    strategy,
		spotGateway: spotGateway,
		accountID:   accountID,
	}
}

// ProcessCandle 处理K线
func (e *Engine) ProcessCandle(ctx context.Context, candle *model.Candle) error {
	signal, err := e.strategy.OnCandle(ctx, candle)
	if err != nil {
		return err
	}

	if signal == nil || signal.Signal == SignalNone {
		return nil
	}

	// 执行交易信号
	return e.executeSignal(ctx, signal)
}

// executeSignal 执行交易信号
func (e *Engine) executeSignal(ctx context.Context, signal *TradeSignal) error {
	if signal.Signal == SignalHold || signal.Signal == SignalNone {
		return nil
	}

	var side model.OrderSide
	if signal.Signal == SignalBuy {
		side = model.OrderSideBuy
	} else {
		side = model.OrderSideSell
	}

	// 下单
	_, err := e.spotGateway.PlaceOrder(ctx, &port.SpotPlaceOrderRequest{
		ClientOrderID: generateOrderID(signal.Symbol),
		Symbol:        signal.Symbol,
		Side:          side,
		Type:          model.OrderTypeMarket,
		Price:         signal.Price,
		Quantity:      signal.Quantity,
	})

	return err
}

// generateOrderID 生成订单ID
func generateOrderID(symbol string) string {
	return uuid.Must(uuid.NewV7()).String() + "-" + symbol
}
