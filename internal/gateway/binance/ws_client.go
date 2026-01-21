package binance

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/iluyuns/alpha-trade/internal/domain/model"
	"github.com/iluyuns/alpha-trade/internal/domain/port"
)

// WSClient Binance WebSocket 客户端 (实现 port.MarketDataRepo)
type WSClient struct {
	baseURL string // wss://stream.binance.com:9443
	client  *SpotClient
	mu      sync.RWMutex
	conns   map[string]*wsConn // stream -> connection
}

// wsConn WebSocket 连接封装
type wsConn struct {
	conn      *websocket.Conn
	mu        sync.Mutex
	ctx       context.Context
	cancel    context.CancelFunc
	reconnect bool
}

// NewWSClient 创建 WebSocket 客户端
func NewWSClient(cfg Config) *WSClient {
	baseURL := "wss://stream.binance.com:9443"
	if cfg.Testnet {
		baseURL = "wss://testnet.binance.vision"
	}

	return &WSClient{
		baseURL: baseURL,
		client:  NewSpotClient(cfg),
		conns:   make(map[string]*wsConn),
	}
}

// SubscribeTicks 订阅 Tick 数据流
func (c *WSClient) SubscribeTicks(ctx context.Context, symbols []string) (<-chan *model.Tick, error) {
	stream := c.buildTickStream(symbols)
	ch := make(chan *model.Tick, 100)

	conn, err := c.dial(ctx, stream)
	if err != nil {
		close(ch)
		return nil, fmt.Errorf("dial websocket failed: %w", err)
	}

	// 启动消息读取协程
	go c.readTickMessages(conn, ch)

	return ch, nil
}

// SubscribeKLines 订阅 K线数据流
func (c *WSClient) SubscribeKLines(ctx context.Context, symbols []string, interval string) (<-chan *model.Candle, error) {
	stream := c.buildKlineStream(symbols, interval)
	ch := make(chan *model.Candle, 100)

	conn, err := c.dial(ctx, stream)
	if err != nil {
		close(ch)
		return nil, fmt.Errorf("dial websocket failed: %w", err)
	}

	// 启动消息读取协程
	go c.readKlineMessages(conn, ch)

	return ch, nil
}

// GetHistoricalKLines 拉取历史 K线
func (c *WSClient) GetHistoricalKLines(ctx context.Context, symbol string, interval string, startTime, endTime int64) ([]*model.Candle, error) {
	// 使用 binance-connector-go 的 KLines API (时间戳转 uint64)
	resp, err := c.client.client.NewKlinesService().
		Symbol(strings.ToUpper(symbol)).
		Interval(interval).
		StartTime(uint64(startTime)).
		EndTime(uint64(endTime)).
		Limit(1000).
		Do(ctx)
	if err != nil {
		return nil, fmt.Errorf("get klines failed: %w", err)
	}

	// 解析响应 (binance-connector-go 返回 []*KlinesResponse)
	candles := make([]*model.Candle, 0, len(resp))
	for _, kline := range resp {
		candle := &model.Candle{
			Symbol:    symbol,
			Interval:  interval,
			Open:      model.MustMoney(kline.Open),
			High:      model.MustMoney(kline.High),
			Low:       model.MustMoney(kline.Low),
			Close:     model.MustMoney(kline.Close),
			Volume:    model.MustMoney(kline.Volume),
			OpenTime:  time.UnixMilli(int64(kline.OpenTime)),
			CloseTime: time.UnixMilli(int64(kline.CloseTime)),
			RecvTime:  time.Now(),
		}
		candles = append(candles, candle)
	}

	return candles, nil
}

// GetLatestPrice 获取最新价格
func (c *WSClient) GetLatestPrice(ctx context.Context, symbol string) (model.Money, error) {
	// 使用 TickerPrice API (返回 []*TickerPriceResponse)
	resp, err := c.client.client.NewTickerPriceService().
		Symbol(strings.ToUpper(symbol)).
		Do(ctx)
	if err != nil {
		return model.Zero(), fmt.Errorf("get ticker price failed: %w", err)
	}

	// 提取价格 (单个 symbol 返回单个元素数组)
	if len(resp) > 0 && resp[0].Price != "" {
		return model.MustMoney(resp[0].Price), nil
	}

	return model.Zero(), fmt.Errorf("empty ticker price response")
}

// dial 建立 WebSocket 连接
func (c *WSClient) dial(ctx context.Context, stream string) (*wsConn, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// 检查是否已存在连接
	if conn, exists := c.conns[stream]; exists {
		return conn, nil
	}

	// 建立新连接
	url := fmt.Sprintf("%s/ws/%s", c.baseURL, stream)
	dialer := websocket.DefaultDialer
	dialer.HandshakeTimeout = 10 * time.Second

	ws, _, err := dialer.DialContext(ctx, url, nil)
	if err != nil {
		return nil, fmt.Errorf("websocket dial failed: %w", err)
	}

	// 设置读写超时
	ws.SetReadDeadline(time.Now().Add(60 * time.Second))
	ws.SetPongHandler(func(string) error {
		ws.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	connCtx, cancel := context.WithCancel(ctx)
	conn := &wsConn{
		conn:      ws,
		ctx:       connCtx,
		cancel:    cancel,
		reconnect: true,
	}

	c.conns[stream] = conn

	// 启动 ping 协程
	go c.ping(conn)

	return conn, nil
}

// ping 心跳维持
func (c *WSClient) ping(conn *wsConn) {
	ticker := time.NewTicker(20 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-conn.ctx.Done():
			return
		case <-ticker.C:
			conn.mu.Lock()
			err := conn.conn.WriteMessage(websocket.PingMessage, nil)
			conn.mu.Unlock()
			if err != nil {
				conn.cancel()
				return
			}
		}
	}
}

// readTickMessages 读取 Tick 消息
func (c *WSClient) readTickMessages(conn *wsConn, ch chan<- *model.Tick) {
	defer close(ch)
	defer conn.conn.Close()

	for {
		select {
		case <-conn.ctx.Done():
			return
		default:
			_, message, err := conn.conn.ReadMessage()
			if err != nil {
				return
			}

			tick := c.parseTickMessage(message)
			if tick != nil {
				select {
				case ch <- tick:
				case <-conn.ctx.Done():
					return
				}
			}
		}
	}
}

// readKlineMessages 读取 K线 消息
func (c *WSClient) readKlineMessages(conn *wsConn, ch chan<- *model.Candle) {
	defer close(ch)
	defer conn.conn.Close()

	for {
		select {
		case <-conn.ctx.Done():
			return
		default:
			_, message, err := conn.conn.ReadMessage()
			if err != nil {
				return
			}

			candle := c.parseKlineMessage(message)
			if candle != nil {
				select {
				case ch <- candle:
				case <-conn.ctx.Done():
					return
				}
			}
		}
	}
}

// parseTickMessage 解析 Tick 消息
func (c *WSClient) parseTickMessage(data []byte) *model.Tick {
	var msg struct {
		EventType string `json:"e"` // "trade"
		EventTime int64  `json:"E"` // Event time
		Symbol    string `json:"s"` // BTCUSDT
		Price     string `json:"p"` // Price
		Quantity  string `json:"q"` // Quantity
	}

	if err := json.Unmarshal(data, &msg); err != nil {
		return nil
	}

	if msg.EventType != "trade" {
		return nil
	}

	return &model.Tick{
		Symbol:    msg.Symbol,
		Price:     model.MustMoney(msg.Price),
		Volume:    model.MustMoney(msg.Quantity),
		EventTime: time.UnixMilli(msg.EventTime),
		RecvTime:  time.Now(),
	}
}

// parseKlineMessage 解析 K线 消息
func (c *WSClient) parseKlineMessage(data []byte) *model.Candle {
	var msg struct {
		EventType string `json:"e"` // "kline"
		EventTime int64  `json:"E"` // Event time
		Symbol    string `json:"s"` // BTCUSDT
		Kline     struct {
			StartTime int64  `json:"t"` // Kline start time
			EndTime   int64  `json:"T"` // Kline close time
			Symbol    string `json:"s"` // Symbol
			Interval  string `json:"i"` // Interval
			Open      string `json:"o"` // Open price
			High      string `json:"h"` // High price
			Low       string `json:"l"` // Low price
			Close     string `json:"c"` // Close price
			Volume    string `json:"v"` // Volume
			IsClosed  bool   `json:"x"` // Is this kline closed?
		} `json:"k"`
	}

	if err := json.Unmarshal(data, &msg); err != nil {
		return nil
	}

	if msg.EventType != "kline" {
		return nil
	}

	// 只返回已关闭的 K线
	if !msg.Kline.IsClosed {
		return nil
	}

	return &model.Candle{
		Symbol:    msg.Symbol,
		Interval:  msg.Kline.Interval,
		Open:      model.MustMoney(msg.Kline.Open),
		High:      model.MustMoney(msg.Kline.High),
		Low:       model.MustMoney(msg.Kline.Low),
		Close:     model.MustMoney(msg.Kline.Close),
		Volume:    model.MustMoney(msg.Kline.Volume),
		OpenTime:  time.UnixMilli(msg.Kline.StartTime),
		CloseTime: time.UnixMilli(msg.Kline.EndTime),
		RecvTime:  time.Now(),
	}
}

// buildTickStream 构建 Tick 流名称
// 单个: btcusdt@trade
// 多个: btcusdt@trade/ethusdt@trade/bnbusdt@trade
func (c *WSClient) buildTickStream(symbols []string) string {
	streams := make([]string, len(symbols))
	for i, symbol := range symbols {
		streams[i] = fmt.Sprintf("%s@trade", strings.ToLower(symbol))
	}
	return strings.Join(streams, "/")
}

// buildKlineStream 构建 K线 流名称
// 单个: btcusdt@kline_1m
// 多个: btcusdt@kline_1m/ethusdt@kline_1m
func (c *WSClient) buildKlineStream(symbols []string, interval string) string {
	streams := make([]string, len(symbols))
	for i, symbol := range symbols {
		streams[i] = fmt.Sprintf("%s@kline_%s", strings.ToLower(symbol), interval)
	}
	return strings.Join(streams, "/")
}

// Close 关闭所有连接
func (c *WSClient) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	for _, conn := range c.conns {
		conn.cancel()
		conn.conn.Close()
	}

	c.conns = make(map[string]*wsConn)
	return nil
}

// 确保 WSClient 实现了 MarketDataRepo 接口
var _ port.MarketDataRepo = (*WSClient)(nil)
