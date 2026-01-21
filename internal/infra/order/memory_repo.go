package order

import (
	"context"
	"fmt"
	"sync"

	"github.com/iluyuns/alpha-trade/internal/domain/model"
)

// MemoryRepo 内存订单仓储（用于回测）
type MemoryRepo struct {
	mu     sync.RWMutex
	orders map[string]*model.Order // key: ClientOrderID
	byExID map[string]string       // key: ExchangeID -> value: ClientOrderID
}

// NewMemoryRepo 创建内存订单仓储
func NewMemoryRepo() *MemoryRepo {
	return &MemoryRepo{
		orders: make(map[string]*model.Order),
		byExID: make(map[string]string),
	}
}

// SaveOrder 保存订单（幂等）
func (r *MemoryRepo) SaveOrder(ctx context.Context, order *model.Order) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// 深拷贝避免外部修改
	copied := copyOrder(order)
	r.orders[order.ClientOrderID] = copied

	// 建立 ExchangeID 索引
	if order.ExchangeID != "" {
		r.byExID[order.ExchangeID] = order.ClientOrderID
	}

	return nil
}

// GetOrder 根据 ClientOrderID 获取订单
func (r *MemoryRepo) GetOrder(ctx context.Context, clientOrderID string) (*model.Order, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	order, exists := r.orders[clientOrderID]
	if !exists {
		return nil, fmt.Errorf("order not found: %s", clientOrderID)
	}

	return copyOrder(order), nil
}

// GetOrderByExchangeID 根据交易所订单ID获取订单
func (r *MemoryRepo) GetOrderByExchangeID(ctx context.Context, exchangeID string) (*model.Order, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	clientOrderID, exists := r.byExID[exchangeID]
	if !exists {
		return nil, fmt.Errorf("order not found by exchange id: %s", exchangeID)
	}

	order, exists := r.orders[clientOrderID]
	if !exists {
		return nil, fmt.Errorf("order not found: %s", clientOrderID)
	}

	return copyOrder(order), nil
}

// UpdateOrderStatus 原子更新订单状态
func (r *MemoryRepo) UpdateOrderStatus(ctx context.Context, clientOrderID string, status model.OrderStatus) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	order, exists := r.orders[clientOrderID]
	if !exists {
		return fmt.Errorf("order not found: %s", clientOrderID)
	}

	order.Status = status
	return nil
}

// UpdateFilled 更新成交数量
func (r *MemoryRepo) UpdateFilled(ctx context.Context, clientOrderID string, filled model.Money) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	order, exists := r.orders[clientOrderID]
	if !exists {
		return fmt.Errorf("order not found: %s", clientOrderID)
	}

	order.Filled = filled
	return nil
}

// ListActiveOrders 列出所有活跃订单
func (r *MemoryRepo) ListActiveOrders(ctx context.Context) ([]*model.Order, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var activeOrders []*model.Order
	for _, order := range r.orders {
		if order.IsActive() {
			activeOrders = append(activeOrders, copyOrder(order))
		}
	}

	return activeOrders, nil
}

// ListOrdersBySymbol 列出指定标的的订单
func (r *MemoryRepo) ListOrdersBySymbol(ctx context.Context, symbol string, limit int) ([]*model.Order, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var orders []*model.Order
	for _, order := range r.orders {
		if order.Symbol == symbol {
			orders = append(orders, copyOrder(order))
			if len(orders) >= limit {
				break
			}
		}
	}

	return orders, nil
}

// copyOrder 深拷贝订单（避免并发修改）
func copyOrder(order *model.Order) *model.Order {
	return &model.Order{
		ClientOrderID: order.ClientOrderID,
		ExchangeID:    order.ExchangeID,
		Symbol:        order.Symbol,
		MarketType:    order.MarketType,
		Side:          order.Side,
		Type:          order.Type,
		Price:         order.Price,
		Quantity:      order.Quantity,
		Filled:        order.Filled,
		Status:        order.Status,
		CreatedAt:     order.CreatedAt,
		UpdatedAt:     order.UpdatedAt,
		SubmitTime:    order.SubmitTime,
		FillTime:      order.FillTime,
		Leverage:      order.Leverage,
		ReduceOnly:    order.ReduceOnly,
		ProtectPrice:  order.ProtectPrice,
	}
}
