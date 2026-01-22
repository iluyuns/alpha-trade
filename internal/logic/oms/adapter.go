package oms

import (
	"context"

	"github.com/iluyuns/alpha-trade/internal/domain/model"
	"github.com/iluyuns/alpha-trade/internal/strategy"
)

// StrategyOMSAdapter OMS 适配器，实现 strategy.OMSInterface
// 用于将 OMS 集成到 Strategy Engine
type StrategyOMSAdapter struct {
	manager *Manager
}

// NewStrategyOMSAdapter 创建策略 OMS 适配器
func NewStrategyOMSAdapter(manager *Manager) *StrategyOMSAdapter {
	return &StrategyOMSAdapter{
		manager: manager,
	}
}

// PlaceOrder 实现 strategy.OMSInterface
func (a *StrategyOMSAdapter) PlaceOrder(ctx context.Context, req *strategy.PlaceOrderRequest) (*model.Order, error) {
	// 转换为 OMS 的 PlaceOrderRequest
	omsReq := &PlaceOrderRequest{
		ClientOrderID: req.ClientOrderID,
		Symbol:        req.Symbol,
		Side:          req.Side,
		Type:          req.Type,
		Price:         req.Price,
		Quantity:      req.Quantity,
		CurrentPrice:  req.CurrentPrice,
		AccountID:     req.AccountID,
		ProtectPrice:  req.ProtectPrice,
	}

	return a.manager.PlaceOrder(ctx, omsReq)
}
