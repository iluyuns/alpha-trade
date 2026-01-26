package oms

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/iluyuns/alpha-trade/internal/domain/model"
	"github.com/iluyuns/alpha-trade/internal/domain/port"
	riskmgr "github.com/iluyuns/alpha-trade/internal/core/risk"
	"github.com/iluyuns/alpha-trade/internal/pkg/metrics"
)

// Manager 订单管理系统（Order Management System）
// 职责：
// 1. 订单状态同步（Gateway <-> OrderRepo）
// 2. 订单生命周期管理
// 3. 与 RiskManager 集成（确保订单通过风控）
type Manager struct {
	mu sync.RWMutex

	// 依赖注入
	spotGateway port.SpotGateway
	orderRepo   port.OrderRepo
	riskMgr     *riskmgr.Manager

	// 配置
	config Config

	// 状态同步
	syncInterval time.Duration // 状态同步间隔
	stopChan     chan struct{}
}

// Config OMS 配置
type Config struct {
	SyncInterval time.Duration // 订单状态同步间隔（默认 5 秒）
	AutoSync     bool          // 是否自动同步订单状态
}

// NewManager 创建订单管理器
func NewManager(
	spotGateway port.SpotGateway,
	orderRepo port.OrderRepo,
	riskMgr *riskmgr.Manager,
	config Config,
) *Manager {
	if config.SyncInterval == 0 {
		config.SyncInterval = 5 * time.Second
	}

	return &Manager{
		spotGateway: spotGateway,
		orderRepo:   orderRepo,
		riskMgr:     riskMgr,
		config:      config,
		stopChan:    make(chan struct{}),
	}
}

// PlaceOrder 下单（集成风控检查）
// 流程：RiskManager.CheckPreTrade -> Gateway.PlaceOrder -> OrderRepo.SaveOrder
func (m *Manager) PlaceOrder(ctx context.Context, req *PlaceOrderRequest) (*model.Order, error) {
	// 1. 风控检查
	orderCtx := &riskmgr.OrderContext{
		ClientOrderID: req.ClientOrderID,
		Symbol:        req.Symbol,
		MarketType:    model.MarketTypeSpot,
		Side:          req.Side,
		Type:          req.Type,
		Price:         req.Price,
		Quantity:      req.Quantity,
		CurrentPrice:  req.CurrentPrice,
		AccountID:     req.AccountID,
		ProtectPrice:  req.ProtectPrice,
	}

	decision, err := m.riskMgr.CheckPreTrade(ctx, orderCtx)
	if err != nil {
		return nil, fmt.Errorf("risk check failed: %w", err)
	}

	if !decision.IsAllowed() {
		return nil, fmt.Errorf("order rejected by risk manager: %s", decision.Reason)
	}

	// 2. 如果风控建议降档，使用建议数量
	quantity := req.Quantity
	if decision.ShouldReduce() && decision.SuggestedQuantity != "" {
		suggestedQty, err := model.NewMoney(decision.SuggestedQuantity)
		if err == nil {
			quantity = suggestedQty
		}
	}

	// 3. 调用 Gateway 下单
	gatewayReq := &port.SpotPlaceOrderRequest{
		ClientOrderID: req.ClientOrderID,
		Symbol:        req.Symbol,
		Side:          req.Side,
		Type:          req.Type,
		Price:         req.Price,
		Quantity:      quantity,
		ProtectPrice:  req.ProtectPrice,
	}

	startTime := time.Now()
	order, err := m.spotGateway.PlaceOrder(ctx, gatewayReq)
	if err != nil {
		metrics.DefaultMetrics.OrdersRejected.Inc()
		return nil, fmt.Errorf("gateway place order failed: %w", err)
	}

	// 记录订单延迟
	metrics.DefaultMetrics.OrderLatency.Observe(time.Since(startTime).Seconds())

	// 4. 持久化订单
	if err := m.orderRepo.SaveOrder(ctx, order); err != nil {
		return nil, fmt.Errorf("save order failed: %w", err)
	}

	// 更新指标
	metrics.DefaultMetrics.OrdersTotal.Inc()
	if order.IsFilled() {
		metrics.DefaultMetrics.OrdersFilled.Inc()
	}

	return order, nil
}

// CancelOrder 撤单
func (m *Manager) CancelOrder(ctx context.Context, clientOrderID string) error {
	// 1. 从 OrderRepo 获取订单
	order, err := m.orderRepo.GetOrder(ctx, clientOrderID)
	if err != nil {
		return fmt.Errorf("get order failed: %w", err)
	}

	// 2. 检查订单状态
	if order.IsClosed() {
		return fmt.Errorf("order already closed: %s", order.Status)
	}

	// 3. 调用 Gateway 撤单
	cancelReq := &port.SpotCancelOrderRequest{
		ClientOrderID: clientOrderID,
		Symbol:        order.Symbol,
	}

	if err := m.spotGateway.CancelOrder(ctx, cancelReq); err != nil {
		return fmt.Errorf("gateway cancel order failed: %w", err)
	}

	// 4. 更新订单状态
	if err := m.orderRepo.UpdateOrderStatus(ctx, clientOrderID, model.OrderStatusCancelled); err != nil {
		return fmt.Errorf("update order status failed: %w", err)
	}

	// 更新指标
	metrics.DefaultMetrics.OrdersCancelled.Inc()

	return nil
}

// SyncOrderStatus 同步订单状态（从 Gateway 同步到 OrderRepo）
func (m *Manager) SyncOrderStatus(ctx context.Context, clientOrderID string) error {
	// 1. 从 Gateway 查询最新状态
	gatewayOrder, err := m.spotGateway.GetOrder(ctx, clientOrderID)
	if err != nil {
		return fmt.Errorf("get order from gateway failed: %w", err)
	}

	// 2. 从 OrderRepo 获取本地状态
	localOrder, err := m.orderRepo.GetOrder(ctx, clientOrderID)
	if err != nil {
		// 本地不存在，直接保存
		return m.orderRepo.SaveOrder(ctx, gatewayOrder)
	}

	// 3. 比较状态，如有变化则更新
	if localOrder.Status != gatewayOrder.Status {
		if err := m.orderRepo.UpdateOrderStatus(ctx, clientOrderID, gatewayOrder.Status); err != nil {
			return fmt.Errorf("update order status failed: %w", err)
		}
	}

	// 4. 更新成交数量
	if !gatewayOrder.Filled.EQ(localOrder.Filled) {
		if err := m.orderRepo.UpdateFilled(ctx, clientOrderID, gatewayOrder.Filled); err != nil {
			return fmt.Errorf("update filled quantity failed: %w", err)
		}
	}

	return nil
}

// SyncActiveOrders 同步所有活跃订单状态
func (m *Manager) SyncActiveOrders(ctx context.Context) error {
	// 1. 获取所有活跃订单
	activeOrders, err := m.orderRepo.ListActiveOrders(ctx)
	if err != nil {
		return fmt.Errorf("list active orders failed: %w", err)
	}

	// 2. 逐个同步
	var lastErr error
	for _, order := range activeOrders {
		if err := m.SyncOrderStatus(ctx, order.ClientOrderID); err != nil {
			lastErr = err
			// 继续同步其他订单，不中断
		}
	}

	return lastErr
}

// StartAutoSync 启动自动同步（后台 goroutine）
func (m *Manager) StartAutoSync(ctx context.Context) {
	if !m.config.AutoSync {
		return
	}

	go func() {
		ticker := time.NewTicker(m.config.SyncInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				_ = m.SyncActiveOrders(ctx)
			case <-m.stopChan:
				return
			case <-ctx.Done():
				return
			}
		}
	}()
}

// StopAutoSync 停止自动同步
func (m *Manager) StopAutoSync() {
	close(m.stopChan)
}

// GetOrder 查询订单（优先从 OrderRepo，不存在则从 Gateway 同步）
func (m *Manager) GetOrder(ctx context.Context, clientOrderID string) (*model.Order, error) {
	// 1. 先查本地
	order, err := m.orderRepo.GetOrder(ctx, clientOrderID)
	if err == nil {
		return order, nil
	}

	// 2. 本地不存在，从 Gateway 查询并保存
	gatewayOrder, err := m.spotGateway.GetOrder(ctx, clientOrderID)
	if err != nil {
		return nil, fmt.Errorf("get order failed: %w", err)
	}

	// 保存到本地
	if err := m.orderRepo.SaveOrder(ctx, gatewayOrder); err != nil {
		return nil, fmt.Errorf("save order failed: %w", err)
	}

	return gatewayOrder, nil
}

// PlaceOrderRequest 下单请求
type PlaceOrderRequest struct {
	ClientOrderID string
	Symbol        string
	Side          model.OrderSide
	Type          model.OrderType
	Price         model.Money
	Quantity      model.Money
	CurrentPrice  model.Money // 当前市价（用于风控计算）
	AccountID     string
	ProtectPrice  model.Money // 保护价
}
