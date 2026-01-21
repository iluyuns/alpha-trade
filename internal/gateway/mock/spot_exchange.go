package mock

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/iluyuns/alpha-trade/internal/domain/model"
	"github.com/iluyuns/alpha-trade/internal/domain/port"
)

// SpotExchange 模拟现货交易所（回测用）
type SpotExchange struct {
	mu sync.RWMutex

	// 订单簿
	orders map[string]*model.Order // key: ClientOrderID

	// 账户余额
	balances map[string]*port.SpotBalance // key: Asset

	// 当前价格（模拟市价）
	currentPrices map[string]model.Money // key: Symbol

	// 配置
	config SpotExchangeConfig
}

// SpotExchangeConfig 模拟交易所配置
type SpotExchangeConfig struct {
	TakerFee    model.Money // 吃单手续费率
	MakerFee    model.Money // 挂单手续费率
	InstantFill bool        // 是否立即成交（回测用）
	Slippage    model.Money // 滑点（百分比）
}

// NewSpotExchange 创建模拟现货交易所
func NewSpotExchange(initialBalance map[string]model.Money) *SpotExchange {
	balances := make(map[string]*port.SpotBalance)
	for asset, amount := range initialBalance {
		balances[asset] = &port.SpotBalance{
			Asset:     asset,
			Free:      amount,
			Locked:    model.Zero(),
			Total:     amount,
			UpdatedAt: time.Now().UnixMilli(),
		}
	}

	return &SpotExchange{
		orders:        make(map[string]*model.Order),
		balances:      balances,
		currentPrices: make(map[string]model.Money),
		config: SpotExchangeConfig{
			TakerFee:    model.MustMoney("0.001"), // 0.1%
			MakerFee:    model.MustMoney("0.001"),
			InstantFill: true,
			Slippage:    model.MustMoney("0.0005"), // 0.05%
		},
	}
}

// PlaceOrder 下单
func (e *SpotExchange) PlaceOrder(ctx context.Context, req *port.SpotPlaceOrderRequest) (*model.Order, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	// 检查订单是否已存在（幂等）
	if existing, exists := e.orders[req.ClientOrderID]; exists {
		return existing, nil
	}

	// 创建订单
	order := &model.Order{
		ClientOrderID: req.ClientOrderID,
		ExchangeID:    fmt.Sprintf("MOCK-%d", time.Now().UnixNano()),
		Symbol:        req.Symbol,
		MarketType:    model.MarketTypeSpot,
		Side:          req.Side,
		Type:          req.Type,
		Price:         req.Price,
		Quantity:      req.Quantity,
		Filled:        model.Zero(),
		Status:        model.OrderStatusPending,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	// 模拟立即成交（回测模式）
	if e.config.InstantFill {
		if err := e.fillOrder(order); err != nil {
			order.Status = model.OrderStatusRejected
			e.orders[req.ClientOrderID] = order
			return order, err
		}
	}

	e.orders[req.ClientOrderID] = order
	return order, nil
}

// fillOrder 模拟订单成交
func (e *SpotExchange) fillOrder(order *model.Order) error {
	// 获取成交价格
	fillPrice := order.Price
	if order.Type == model.OrderTypeMarket {
		price, exists := e.currentPrices[order.Symbol]
		if !exists {
			return fmt.Errorf("no market price for %s", order.Symbol)
		}
		// 模拟滑点
		slippage := price.Mul(e.config.Slippage)
		if order.Side == model.OrderSideBuy {
			fillPrice = price.Add(slippage)
		} else {
			fillPrice = price.Sub(slippage)
		}
	}

	// 解析交易对（简化：假设 BTCUSDT 格式）
	baseAsset, quoteAsset := parseSymbol(order.Symbol)

	if order.Side == model.OrderSideBuy {
		// 买入：扣除报价资产，增加基础资产
		cost := fillPrice.Mul(order.Quantity)
		fee := cost.Mul(e.config.TakerFee)
		totalCost := cost.Add(fee)

		quoteBal, exists := e.balances[quoteAsset]
		if !exists || quoteBal.Free.LT(totalCost) {
			return fmt.Errorf("insufficient %s balance", quoteAsset)
		}

		quoteBal.Free = quoteBal.Free.Sub(totalCost)
		quoteBal.Total = quoteBal.Free.Add(quoteBal.Locked)

		baseBal := e.getOrCreateBalance(baseAsset)
		baseBal.Free = baseBal.Free.Add(order.Quantity)
		baseBal.Total = baseBal.Free.Add(baseBal.Locked)

	} else {
		// 卖出：扣除基础资产，增加报价资产
		baseBal, exists := e.balances[baseAsset]
		if !exists || baseBal.Free.LT(order.Quantity) {
			return fmt.Errorf("insufficient %s balance", baseAsset)
		}

		baseBal.Free = baseBal.Free.Sub(order.Quantity)
		baseBal.Total = baseBal.Free.Add(baseBal.Locked)

		revenue := fillPrice.Mul(order.Quantity)
		fee := revenue.Mul(e.config.TakerFee)
		netRevenue := revenue.Sub(fee)

		quoteBal := e.getOrCreateBalance(quoteAsset)
		quoteBal.Free = quoteBal.Free.Add(netRevenue)
		quoteBal.Total = quoteBal.Free.Add(quoteBal.Locked)
	}

	// 更新订单状态
	order.Filled = order.Quantity
	order.Status = model.OrderStatusFilled
	order.SubmitTime = time.Now()
	order.FillTime = time.Now()
	order.UpdatedAt = time.Now()

	return nil
}

// CancelOrder 撤单
func (e *SpotExchange) CancelOrder(ctx context.Context, req *port.SpotCancelOrderRequest) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	order, exists := e.orders[req.ClientOrderID]
	if !exists {
		return fmt.Errorf("order %s not found", req.ClientOrderID)
	}

	if order.IsClosed() {
		return fmt.Errorf("order already closed")
	}

	order.Status = model.OrderStatusCancelled
	order.UpdatedAt = time.Now()
	return nil
}

// GetOrder 查询订单
func (e *SpotExchange) GetOrder(ctx context.Context, clientOrderID string) (*model.Order, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	order, exists := e.orders[clientOrderID]
	if !exists {
		return nil, fmt.Errorf("order %s not found", clientOrderID)
	}

	return order, nil
}

// GetBalance 查询余额
func (e *SpotExchange) GetBalance(ctx context.Context, asset string) (*port.SpotBalance, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	balance, exists := e.balances[asset]
	if !exists {
		return &port.SpotBalance{
			Asset:     asset,
			Free:      model.Zero(),
			Locked:    model.Zero(),
			Total:     model.Zero(),
			UpdatedAt: time.Now().UnixMilli(),
		}, nil
	}

	return balance, nil
}

// GetAllBalances 查询所有余额
func (e *SpotExchange) GetAllBalances(ctx context.Context) ([]*port.SpotBalance, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	balances := make([]*port.SpotBalance, 0, len(e.balances))
	for _, bal := range e.balances {
		balances = append(balances, bal)
	}

	return balances, nil
}

// SetPrice 设置当前价格（回测用）
func (e *SpotExchange) SetPrice(symbol string, price model.Money) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.currentPrices[symbol] = price
}

// getOrCreateBalance 获取或创建余额
func (e *SpotExchange) getOrCreateBalance(asset string) *port.SpotBalance {
	if bal, exists := e.balances[asset]; exists {
		return bal
	}

	bal := &port.SpotBalance{
		Asset:     asset,
		Free:      model.Zero(),
		Locked:    model.Zero(),
		Total:     model.Zero(),
		UpdatedAt: time.Now().UnixMilli(),
	}
	e.balances[asset] = bal
	return bal
}

// parseSymbol 解析交易对符号（简化实现）
func parseSymbol(symbol string) (base, quote string) {
	// 简化：假设 USDT 结尾
	if len(symbol) > 4 && symbol[len(symbol)-4:] == "USDT" {
		return symbol[:len(symbol)-4], "USDT"
	}
	// 默认
	return symbol[:3], symbol[3:]
}
