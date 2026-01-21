package model

import "time"

// OrderSide 订单方向
type OrderSide int

const (
	OrderSideBuy OrderSide = iota + 1
	OrderSideSell
)

func (s OrderSide) String() string {
	switch s {
	case OrderSideBuy:
		return "BUY"
	case OrderSideSell:
		return "SELL"
	default:
		return "UNKNOWN"
	}
}

// OrderType 订单类型
type OrderType int

const (
	OrderTypeLimit OrderType = iota + 1
	OrderTypeMarket
	OrderTypeIOC // Immediate-Or-Cancel
	OrderTypeFOK // Fill-Or-Kill
)

func (t OrderType) String() string {
	switch t {
	case OrderTypeLimit:
		return "LIMIT"
	case OrderTypeMarket:
		return "MARKET"
	case OrderTypeIOC:
		return "IOC"
	case OrderTypeFOK:
		return "FOK"
	default:
		return "UNKNOWN"
	}
}

// OrderStatus 订单状态
type OrderStatus int

const (
	OrderStatusPending OrderStatus = iota + 1
	OrderStatusSubmitted
	OrderStatusPartialFilled
	OrderStatusFilled
	OrderStatusCancelled
	OrderStatusRejected
)

func (s OrderStatus) String() string {
	switch s {
	case OrderStatusPending:
		return "PENDING"
	case OrderStatusSubmitted:
		return "SUBMITTED"
	case OrderStatusPartialFilled:
		return "PARTIAL_FILLED"
	case OrderStatusFilled:
		return "FILLED"
	case OrderStatusCancelled:
		return "CANCELLED"
	case OrderStatusRejected:
		return "REJECTED"
	default:
		return "UNKNOWN"
	}
}

// MarketType 市场类型
type MarketType int

const (
	MarketTypeSpot MarketType = iota + 1
	MarketTypeFuture
)

func (m MarketType) String() string {
	switch m {
	case MarketTypeSpot:
		return "SPOT"
	case MarketTypeFuture:
		return "FUTURE"
	default:
		return "UNKNOWN"
	}
}

// Order 订单领域模型
type Order struct {
	// 唯一标识
	ClientOrderID string // 客户端订单ID（幂等保证）
	ExchangeID    string // 交易所订单ID

	// 基础属性
	Symbol     string
	MarketType MarketType
	Side       OrderSide
	Type       OrderType

	// 价格与数量
	Price    Money // 限价单价格（市价单为零）
	Quantity Money // 下单数量
	Filled   Money // 已成交数量

	// 状态
	Status OrderStatus

	// 时间
	CreatedAt  time.Time
	UpdatedAt  time.Time
	SubmitTime time.Time // 提交到交易所时间
	FillTime   time.Time // 完全成交时间

	// 合约专属（现货为零值）
	Leverage   int  // 杠杆倍数
	ReduceOnly bool // 只减仓

	// 保护价（风控用）
	ProtectPrice Money // 最差成交价格
}

// IsFilled 是否完全成交
func (o *Order) IsFilled() bool {
	return o.Status == OrderStatusFilled
}

// IsActive 是否活跃（未完结）
func (o *Order) IsActive() bool {
	return o.Status == OrderStatusPending ||
		o.Status == OrderStatusSubmitted ||
		o.Status == OrderStatusPartialFilled
}

// IsClosed 是否已关闭
func (o *Order) IsClosed() bool {
	return o.Status == OrderStatusFilled ||
		o.Status == OrderStatusCancelled ||
		o.Status == OrderStatusRejected
}

// FilledPercent 成交百分比
func (o *Order) FilledPercent() Money {
	if o.Quantity.IsZero() {
		return Zero()
	}
	return o.Filled.Div(o.Quantity)
}

// RemainingQty 剩余数量
func (o *Order) RemainingQty() Money {
	return o.Quantity.Sub(o.Filled)
}
