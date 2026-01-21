package port

import (
	"context"

	"github.com/iluyuns/alpha-trade/internal/domain/model"
)

// SpotPlaceOrderRequest 现货下单请求
type SpotPlaceOrderRequest struct {
	ClientOrderID string          // 客户端订单ID（幂等）
	Symbol        string          // 交易对
	Side          model.OrderSide // 买卖方向
	Type          model.OrderType // 订单类型
	Price         model.Money     // 限价单价格
	Quantity      model.Money     // 数量
	ProtectPrice  model.Money     // 保护价（最差成交价）
}

// SpotCancelOrderRequest 撤单请求
type SpotCancelOrderRequest struct {
	ClientOrderID string // 客户端订单ID
	ExchangeID    string // 交易所订单ID（二选一）
	Symbol        string // 交易对
}

// SpotBalance 现货余额
type SpotBalance struct {
	Asset     string
	Free      model.Money // 可用余额
	Locked    model.Money // 冻结余额
	Total     model.Money // 总余额
	UpdatedAt int64       // 更新时间（Unix毫秒）
}

// SpotGateway 现货交易接口
type SpotGateway interface {
	// PlaceOrder 下单
	PlaceOrder(ctx context.Context, req *SpotPlaceOrderRequest) (*model.Order, error)

	// CancelOrder 撤单
	CancelOrder(ctx context.Context, req *SpotCancelOrderRequest) error

	// GetOrder 查询订单
	GetOrder(ctx context.Context, clientOrderID string) (*model.Order, error)

	// GetBalance 查询余额
	GetBalance(ctx context.Context, asset string) (*SpotBalance, error)

	// GetAllBalances 查询所有余额
	GetAllBalances(ctx context.Context) ([]*SpotBalance, error)
}
