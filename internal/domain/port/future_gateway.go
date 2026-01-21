package port

import (
	"context"

	"github.com/iluyuns/alpha-trade/internal/domain/model"
)

// FuturePlaceOrderRequest 合约下单请求
type FuturePlaceOrderRequest struct {
	ClientOrderID string          // 客户端订单ID（幂等）
	Symbol        string          // 合约交易对
	Side          model.OrderSide // 买卖方向
	Type          model.OrderType // 订单类型
	Price         model.Money     // 限价单价格
	Quantity      model.Money     // 数量（张数）
	Leverage      int             // 杠杆倍数
	ReduceOnly    bool            // 只减仓
	ProtectPrice  model.Money     // 保护价（最差成交价）
}

// FutureCancelOrderRequest 撤单请求
type FutureCancelOrderRequest struct {
	ClientOrderID string // 客户端订单ID
	ExchangeID    string // 交易所订单ID（二选一）
	Symbol        string // 合约交易对
}

// FuturePosition 合约持仓
type FuturePosition struct {
	Symbol           string
	Side             model.OrderSide // LONG/SHORT
	Size             model.Money     // 持仓数量
	EntryPrice       model.Money     // 开仓均价
	MarkPrice        model.Money     // 标记价格
	Leverage         int             // 当前杠杆
	UnrealizedPnL    model.Money     // 未实现盈亏
	LiquidationPrice model.Money     // 强平价
	UpdatedAt        int64           // 更新时间（Unix毫秒）
}

// FutureBalance 合约账户余额
type FutureBalance struct {
	Asset             string
	WalletBalance     model.Money // 钱包余额
	UnrealizedPnL     model.Money // 未实现盈亏
	MarginBalance     model.Money // 保证金余额
	AvailableBalance  model.Money // 可用余额
	MaxWithdrawAmount model.Money // 最大可提现金额
	UpdatedAt         int64       // 更新时间（Unix毫秒）
}

// FutureGateway 合约交易接口
type FutureGateway interface {
	// PlaceOrder 下单
	PlaceOrder(ctx context.Context, req *FuturePlaceOrderRequest) (*model.Order, error)

	// CancelOrder 撤单
	CancelOrder(ctx context.Context, req *FutureCancelOrderRequest) error

	// GetOrder 查询订单
	GetOrder(ctx context.Context, clientOrderID string) (*model.Order, error)

	// GetPosition 查询持仓
	GetPosition(ctx context.Context, symbol string) (*FuturePosition, error)

	// GetAllPositions 查询所有持仓
	GetAllPositions(ctx context.Context) ([]*FuturePosition, error)

	// GetBalance 查询账户余额
	GetBalance(ctx context.Context) (*FutureBalance, error)

	// SetLeverage 设置杠杆
	SetLeverage(ctx context.Context, symbol string, leverage int) error
}
