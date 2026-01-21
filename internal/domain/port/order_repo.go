package port

import (
	"context"

	"github.com/iluyuns/alpha-trade/internal/domain/model"
)

// OrderRepo 订单持久化接口
// 实现要求：
// 1. 幂等写入（相同 ClientOrderID 多次写入结果一致）
// 2. 原子更新（状态变更全部成功或全部失败）
// 3. 支持回测（内存实现）与实盘（DB 实现）
type OrderRepo interface {
	// SaveOrder 保存订单（幂等）
	// 如果 ClientOrderID 已存在则更新，否则插入
	SaveOrder(ctx context.Context, order *model.Order) error

	// GetOrder 根据 ClientOrderID 获取订单
	GetOrder(ctx context.Context, clientOrderID string) (*model.Order, error)

	// GetOrderByExchangeID 根据交易所订单ID获取订单
	GetOrderByExchangeID(ctx context.Context, exchangeID string) (*model.Order, error)

	// UpdateOrderStatus 原子更新订单状态
	UpdateOrderStatus(ctx context.Context, clientOrderID string, status model.OrderStatus) error

	// UpdateFilled 更新成交数量
	UpdateFilled(ctx context.Context, clientOrderID string, filled model.Money) error

	// ListActiveOrders 列出所有活跃订单
	ListActiveOrders(ctx context.Context) ([]*model.Order, error)

	// ListOrdersBySymbol 列出指定标的的订单
	ListOrdersBySymbol(ctx context.Context, symbol string, limit int) ([]*model.Order, error)
}
