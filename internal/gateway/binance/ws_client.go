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
	"github.com/zeromicro/go-zero/core/logx"
)

const (
	maxReconnectAttempts = 10
	initialReconnectWait = 1 * time.Second
	maxReconnectWait     = 60 * time.Second
	pingInterval         = 20 * time.Second
	readTimeout          = 60 * time.Second
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
	conn   *websocket.Conn
	mu     sync.Mutex
	ctx    context.Context
	cancel context.CancelFunc
	stream string
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

	go c.readTickMessagesWithReconnect(conn, ch)

	return ch, nil
}

// SubscribeKLines 订阅 K线数据流
func (c *WSClient) SubscribeKLines(ctx context.Context, symbols []string, interval string) (<-chan *model.Candle, error) {
	stream := c.buildKlineStream(symbols, interval)
	ch := make(chan *model.Candle, 100)

	conn, err := c.dial(ctx, stream)
	if err != nil {
		close(ch)
		return nil, fmt.Errorf("subscribe klines failed: %w", err)
	}

	go c.readKlineMessagesWithReconnect(conn, ch)

	return ch, nil
}

// GetHistoricalKLines 拉取历史 K线
func (c *WSClient) GetHistoricalKLines(ctx context.Context, symbol string, interval string, startTime, endTime int64) ([]*model.Candle, error) {
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
	resp, err := c.client.client.NewTickerPriceService().
		Symbol(strings.ToUpper(symbol)).
		Do(ctx)
	if err != nil {
		return model.Zero(), fmt.Errorf("get ticker price failed: %w", err)
	}

	if len(resp) > 0 && resp[0].Price != "" {
		return model.MustMoney(resp[0].Price), nil
	}

	return model.Zero(), fmt.Errorf("empty ticker price response")
}

// dial 建立 WebSocket 连接
func (c *WSClient) dial(ctx context.Context, stream string) (*wsConn, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if conn, exists := c.conns[stream]; exists {
		return conn, nil
	}

	ws, err := c.connectWS(ctx, stream)
	if err != nil {
		return nil, err
	}

	connCtx, cancel := context.WithCancel(ctx)
	conn := &wsConn{
		conn:   ws,
		ctx:    connCtx,
		cancel: cancel,
		stream: stream,
	}

	c.conns[stream] = conn
	go c.ping(conn)

	return conn, nil
}

// connectWS 建立原始 WebSocket 连接
func (c *WSClient) connectWS(ctx context.Context, stream string) (*websocket.Conn, error) {
	url := fmt.Sprintf("%s/ws/%s", c.baseURL, stream)
	dialer := websocket.DefaultDialer
	dialer.HandshakeTimeout = 10 * time.Second

	ws, _, err := dialer.DialContext(ctx, url, nil)
	if err != nil {
		return nil, fmt.Errorf("websocket dial failed: %w", err)
	}

	ws.SetReadDeadline(time.Now().Add(readTimeout))
	ws.SetPongHandler(func(string) error {
		ws.SetReadDeadline(time.Now().Add(readTimeout))
		return nil
	})

	return ws, nil
}

// reconnect 指数退避重连
func (c *WSClient) reconnect(conn *wsConn) (*websocket.Conn, error) {
	wait := initialReconnectWait
	for attempt := 1; attempt <= maxReconnectAttempts; attempt++ {
		select {
		case <-conn.ctx.Done():
			return nil, conn.ctx.Err()
		default:
		}

		logx.Infof("[WS] Reconnecting stream=%s attempt=%d/%d", conn.stream, attempt, maxReconnectAttempts)
		time.Sleep(wait)

		ws, err := c.connectWS(conn.ctx, conn.stream)
		if err == nil {
			logx.Infof("[WS] Reconnected stream=%s", conn.stream)

			c.mu.Lock()
			conn.mu.Lock()
			conn.conn = ws
			conn.mu.Unlock()
			c.mu.Unlock()

			go c.ping(conn)
			return ws, nil
		}

		logx.Errorf("[WS] Reconnect failed stream=%s: %v", conn.stream, err)
		wait = wait * 2
		if wait > maxReconnectWait {
			wait = maxReconnectWait
		}
	}

	return nil, fmt.Errorf("max reconnect attempts (%d) exceeded for stream %s", maxReconnectAttempts, conn.stream)
}

// ping 心跳维持
func (c *WSClient) ping(conn *wsConn) {
	ticker := time.NewTicker(pingInterval)
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
				return
			}
		}
	}
}

// readTickMessagesWithReconnect 带自动重连的 Tick 读取循环
func (c *WSClient) readTickMessagesWithReconnect(conn *wsConn, ch chan<- *model.Tick) {
	defer close(ch)

	for {
		select {
		case <-conn.ctx.Done():
			return
		default:
		}

		_, message, err := conn.conn.ReadMessage()
		if err != nil {
			select {
			case <-conn.ctx.Done():
				return
			default:
			}

			conn.conn.Close()
			if _, reconnErr := c.reconnect(conn); reconnErr != nil {
				logx.Errorf("[WS] Tick stream permanently disconnected: %v", reconnErr)
				return
			}
			continue
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

// readKlineMessagesWithReconnect 带自动重连的 K线 读取循环
func (c *WSClient) readKlineMessagesWithReconnect(conn *wsConn, ch chan<- *model.Candle) {
	defer close(ch)

	for {
		select {
		case <-conn.ctx.Done():
			return
		default:
		}

		_, message, err := conn.conn.ReadMessage()
		if err != nil {
			select {
			case <-conn.ctx.Done():
				return
			default:
			}

			conn.conn.Close()
			if _, reconnErr := c.reconnect(conn); reconnErr != nil {
				logx.Errorf("[WS] Kline stream permanently disconnected: %v", reconnErr)
				return
			}
			continue
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

// parseTickMessage 解析 Tick 消息
func (c *WSClient) parseTickMessage(data []byte) *model.Tick {
	var msg struct {
		EventType string `json:"e"` // "trade"
		EventTime int64  `json:"E"`
		Symbol    string `json:"s"`
		Price     string `json:"p"`
		Quantity  string `json:"q"`
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
		EventTime int64  `json:"E"`
		Symbol    string `json:"s"`
		Kline     struct {
			StartTime int64  `json:"t"`
			EndTime   int64  `json:"T"`
			Symbol    string `json:"s"`
			Interval  string `json:"i"`
			Open      string `json:"o"`
			High      string `json:"h"`
			Low       string `json:"l"`
			Close     string `json:"c"`
			Volume    string `json:"v"`
			IsClosed  bool   `json:"x"`
		} `json:"k"`
	}

	if err := json.Unmarshal(data, &msg); err != nil {
		return nil
	}

	if msg.EventType != "kline" {
		return nil
	}

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
func (c *WSClient) buildTickStream(symbols []string) string {
	streams := make([]string, len(symbols))
	for i, symbol := range symbols {
		streams[i] = fmt.Sprintf("%s@trade", strings.ToLower(symbol))
	}
	return strings.Join(streams, "/")
}

// buildKlineStream 构建 K线 流名称
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

var _ port.MarketDataRepo = (*WSClient)(nil)
