package binance

import (
	"context"
	"fmt"
	"strconv"
	"time"

	binance_connector "github.com/binance/binance-connector-go"
	"github.com/iluyuns/alpha-trade/internal/domain/model"
	"github.com/iluyuns/alpha-trade/internal/domain/port"
)

// SpotClient Binance 现货客户端（实现 port.SpotGateway 接口）
type SpotClient struct {
	client    *binance_connector.Client
	apiKey    string
	apiSecret string
}

// Config Binance 客户端配置
type Config struct {
	APIKey    string
	APISecret string
	BaseURL   string // 默认: https://api.binance.com
	Testnet   bool   // 是否使用测试网
}

// NewSpotClient 创建 Binance 现货客户端
func NewSpotClient(cfg Config) *SpotClient {
	baseURL := cfg.BaseURL
	if cfg.Testnet {
		baseURL = "https://testnet.binance.vision"
	} else if baseURL == "" {
		baseURL = "https://api.binance.com"
	}

	client := binance_connector.NewClient(cfg.APIKey, cfg.APISecret, baseURL)

	return &SpotClient{
		client:    client,
		apiKey:    cfg.APIKey,
		apiSecret: cfg.APISecret,
	}
}

// PlaceOrder 下单
func (c *SpotClient) PlaceOrder(ctx context.Context, req *port.SpotPlaceOrderRequest) (*model.Order, error) {
	// 转换订单类型和方向
	orderType := c.convertOrderType(req.Type)
	side := c.convertSide(req.Side)

	// 构建请求
	builder := c.client.NewCreateOrderService().
		Symbol(req.Symbol).
		Side(side).
		Type(orderType).
		NewClientOrderId(req.ClientOrderID)

	// 数量（转换为 float64）
	quantity, _ := strconv.ParseFloat(req.Quantity.String(), 64)
	builder = builder.Quantity(quantity)

	// 限价单需要价格
	if req.Type == model.OrderTypeLimit {
		price, _ := strconv.ParseFloat(req.Price.String(), 64)
		builder = builder.Price(price).
			TimeInForce("GTC") // Good Till Cancel
	}

	// 执行下单
	resp, err := builder.Do(ctx)
	if err != nil {
		return nil, fmt.Errorf("binance place order failed: %w", err)
	}

	// 转换响应
	return c.convertOrderResponse(resp), nil
}

// CancelOrder 撤单
func (c *SpotClient) CancelOrder(ctx context.Context, req *port.SpotCancelOrderRequest) error {
	builder := c.client.NewCancelOrderService().
		Symbol(req.Symbol)

	if req.ClientOrderID != "" {
		builder = builder.OrigClientOrderId(req.ClientOrderID)
	} else if req.ExchangeID != "" {
		orderID, _ := strconv.ParseInt(req.ExchangeID, 10, 64)
		builder = builder.OrderId(orderID)
	} else {
		return fmt.Errorf("either ClientOrderID or ExchangeID must be provided")
	}

	_, err := builder.Do(ctx)
	if err != nil {
		return fmt.Errorf("binance cancel order failed: %w", err)
	}

	return nil
}

// GetOrder 查询订单
func (c *SpotClient) GetOrder(ctx context.Context, clientOrderID string) (*model.Order, error) {
	// Binance 需要 symbol，但我们的接口设计只传 clientOrderID
	// 这是一个设计缺陷，需要改进接口或在订单ID中编码symbol
	// 临时方案：从 clientOrderID 解析 symbol（假设格式: SYMBOL-UUID）
	symbol := c.extractSymbolFromOrderID(clientOrderID)
	if symbol == "" {
		return nil, fmt.Errorf("cannot extract symbol from clientOrderID: %s", clientOrderID)
	}

	resp, err := c.client.NewGetOrderService().
		Symbol(symbol).
		OrigClientOrderId(clientOrderID).
		Do(ctx)
	if err != nil {
		return nil, fmt.Errorf("binance get order failed: %w", err)
	}

	return c.convertOrderResponse(resp), nil
}

// GetBalance 查询余额
func (c *SpotClient) GetBalance(ctx context.Context, asset string) (*port.SpotBalance, error) {
	resp, err := c.client.NewGetAccountService().Do(ctx)
	if err != nil {
		return nil, fmt.Errorf("binance get account failed: %w", err)
	}

	// 查找指定资产
	for _, balance := range resp.Balances {
		if balance.Asset == asset {
			free := model.MustMoney(balance.Free)
			locked := model.MustMoney(balance.Locked)
			return &port.SpotBalance{
				Asset:     asset,
				Free:      free,
				Locked:    locked,
				Total:     free.Add(locked),
				UpdatedAt: time.Now().UnixMilli(),
			}, nil
		}
	}

	// 资产不存在，返回零余额
	return &port.SpotBalance{
		Asset:     asset,
		Free:      model.Zero(),
		Locked:    model.Zero(),
		Total:     model.Zero(),
		UpdatedAt: time.Now().UnixMilli(),
	}, nil
}

// GetAllBalances 查询所有余额
func (c *SpotClient) GetAllBalances(ctx context.Context) ([]*port.SpotBalance, error) {
	resp, err := c.client.NewGetAccountService().Do(ctx)
	if err != nil {
		return nil, fmt.Errorf("binance get account failed: %w", err)
	}

	balances := make([]*port.SpotBalance, 0, len(resp.Balances))
	now := time.Now().UnixMilli()

	for _, b := range resp.Balances {
		free := model.MustMoney(b.Free)
		locked := model.MustMoney(b.Locked)
		total := free.Add(locked)

		// 只返回非零余额
		if !total.IsZero() {
			balances = append(balances, &port.SpotBalance{
				Asset:     b.Asset,
				Free:      free,
				Locked:    locked,
				Total:     total,
				UpdatedAt: now,
			})
		}
	}

	return balances, nil
}

// convertOrderType 转换订单类型
func (c *SpotClient) convertOrderType(t model.OrderType) string {
	switch t {
	case model.OrderTypeMarket:
		return "MARKET"
	case model.OrderTypeLimit:
		return "LIMIT"
	case model.OrderTypeIOC:
		return "LIMIT" // IOC 通过 TimeInForce 实现
	case model.OrderTypeFOK:
		return "LIMIT" // FOK 通过 TimeInForce 实现
	default:
		return "LIMIT"
	}
}

// convertSide 转换买卖方向
func (c *SpotClient) convertSide(side model.OrderSide) string {
	if side == model.OrderSideBuy {
		return "BUY"
	}
	return "SELL"
}

// convertOrderResponse 转换订单响应
func (c *SpotClient) convertOrderResponse(data interface{}) *model.Order {
	// binance_connector 返回的是结构化数据
	// 使用类型断言获取字段
	var orderID int64
	var clientOrderID, symbol, status, side, price, origQty, executedQty string

	// 使用反射或 map 方式解析（这里简化为 map）
	if orderMap, ok := data.(map[string]interface{}); ok {
		if oid, ok := orderMap["orderId"].(float64); ok {
			orderID = int64(oid)
		}
		clientOrderID, _ = orderMap["clientOrderId"].(string)
		symbol, _ = orderMap["symbol"].(string)
		status, _ = orderMap["status"].(string)
		side, _ = orderMap["side"].(string)
		price, _ = orderMap["price"].(string)
		origQty, _ = orderMap["origQty"].(string)
		executedQty, _ = orderMap["executedQty"].(string)
	}

	order := &model.Order{
		ClientOrderID: clientOrderID,
		ExchangeID:    strconv.FormatInt(orderID, 10),
		Symbol:        symbol,
		MarketType:    model.MarketTypeSpot,
		Status:        c.convertOrderStatus(status),
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	// 解析买卖方向
	if side == "BUY" {
		order.Side = model.OrderSideBuy
	} else {
		order.Side = model.OrderSideSell
	}

	// 解析数量和价格
	if price != "" {
		order.Price = model.MustMoney(price)
	}
	if origQty != "" {
		order.Quantity = model.MustMoney(origQty)
	}
	if executedQty != "" {
		order.Filled = model.MustMoney(executedQty)
	}

	return order
}

// convertOrderStatus 转换订单状态
func (c *SpotClient) convertOrderStatus(status string) model.OrderStatus {
	switch status {
	case "NEW":
		return model.OrderStatusSubmitted
	case "PARTIALLY_FILLED":
		return model.OrderStatusPartialFilled
	case "FILLED":
		return model.OrderStatusFilled
	case "CANCELED":
		return model.OrderStatusCancelled
	case "REJECTED":
		return model.OrderStatusRejected
	case "EXPIRED":
		return model.OrderStatusCancelled
	default:
		return model.OrderStatusPending
	}
}

// extractSymbolFromOrderID 从订单ID提取交易对
// 假设订单ID格式: BTCUSDT-uuid 或 uuid-BTCUSDT
func (c *SpotClient) extractSymbolFromOrderID(orderID string) string {
	// 简化实现：假设前缀是 symbol
	// 实际应该有更健壮的解析逻辑
	if len(orderID) > 7 && orderID[0:3] != "" {
		// 尝试提取前缀
		for i := 0; i < len(orderID); i++ {
			if orderID[i] == '-' {
				return orderID[0:i]
			}
		}
	}
	return ""
}

