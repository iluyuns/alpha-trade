package binance

import (
	"context"
	"os"
	"testing"
	"time"
)

// TestWSClient_SubscribeTicks 测试订阅 Tick 数据
func TestWSClient_SubscribeTicks(t *testing.T) {
	apiKey := os.Getenv("BINANCE_API_KEY")
	apiSecret := os.Getenv("BINANCE_API_SECRET")

	if apiKey == "" || apiSecret == "" {
		t.Skip("Skipping integration test: BINANCE_API_KEY or BINANCE_API_SECRET not set")
	}

	client := NewWSClient(Config{
		APIKey:    apiKey,
		APISecret: apiSecret,
		Testnet:   true,
	})
	defer client.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	symbols := []string{"BTCUSDT", "ETHUSDT"}
	tickCh, err := client.SubscribeTicks(ctx, symbols)
	if err != nil {
		t.Fatalf("SubscribeTicks failed: %v", err)
	}

	// 读取至少 5 个 tick
	count := 0
	for tick := range tickCh {
		t.Logf("Tick: %s @ %s (Volume: %s, Time: %s)",
			tick.Symbol, tick.Price.String(), tick.Volume.String(), tick.EventTime)
		count++
		if count >= 5 {
			cancel()
			break
		}
	}

	if count == 0 {
		t.Error("No ticks received")
	}
}

// TestWSClient_SubscribeKLines 测试订阅 K线数据
func TestWSClient_SubscribeKLines(t *testing.T) {
	apiKey := os.Getenv("BINANCE_API_KEY")
	apiSecret := os.Getenv("BINANCE_API_SECRET")

	if apiKey == "" || apiSecret == "" {
		t.Skip("Skipping integration test: BINANCE_API_KEY or BINANCE_API_SECRET not set")
	}

	client := NewWSClient(Config{
		APIKey:    apiKey,
		APISecret: apiSecret,
		Testnet:   true,
	})
	defer client.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	symbols := []string{"BTCUSDT"}
	interval := "1m"
	klineCh, err := client.SubscribeKLines(ctx, symbols, interval)
	if err != nil {
		t.Fatalf("SubscribeKLines failed: %v", err)
	}

	// 等待至少 1 个 K线（可能需要等到下一分钟）
	count := 0
	for candle := range klineCh {
		t.Logf("Candle: %s [%s] O:%s H:%s L:%s C:%s V:%s (Time: %s)",
			candle.Symbol, candle.Interval,
			candle.Open.String(), candle.High.String(), candle.Low.String(),
			candle.Close.String(), candle.Volume.String(), candle.OpenTime)
		count++
		if count >= 1 {
			cancel()
			break
		}
	}

	if count == 0 {
		t.Error("No klines received")
	}
}

// TestBuildTickStream 测试构建 Tick 流名称
func TestBuildTickStream(t *testing.T) {
	client := NewWSClient(Config{})

	tests := []struct {
		name    string
		symbols []string
		want    string
	}{
		{
			name:    "single symbol",
			symbols: []string{"BTCUSDT"},
			want:    "btcusdt@trade",
		},
		{
			name:    "multiple symbols",
			symbols: []string{"BTCUSDT", "ETHUSDT"},
			want:    "btcusdt@trade/ethusdt@trade",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := client.buildTickStream(tt.symbols)
			if got != tt.want {
				t.Errorf("buildTickStream() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestBuildKlineStream 测试构建 K线 流名称
func TestBuildKlineStream(t *testing.T) {
	client := NewWSClient(Config{})

	tests := []struct {
		name     string
		symbols  []string
		interval string
		want     string
	}{
		{
			name:     "single symbol 1m",
			symbols:  []string{"BTCUSDT"},
			interval: "1m",
			want:     "btcusdt@kline_1m",
		},
		{
			name:     "multiple symbols 5m",
			symbols:  []string{"BTCUSDT", "ETHUSDT"},
			interval: "5m",
			want:     "btcusdt@kline_5m/ethusdt@kline_5m",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := client.buildKlineStream(tt.symbols, tt.interval)
			if got != tt.want {
				t.Errorf("buildKlineStream() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestParseTickMessage 测试解析 Tick 消息
func TestParseTickMessage(t *testing.T) {
	client := NewWSClient(Config{})

	// Binance 真实消息格式
	data := []byte(`{
		"e": "trade",
		"E": 1672531200000,
		"s": "BTCUSDT",
		"p": "16500.50",
		"q": "0.123"
	}`)

	tick := client.parseTickMessage(data)
	if tick == nil {
		t.Fatal("parseTickMessage returned nil")
	}

	if tick.Symbol != "BTCUSDT" {
		t.Errorf("Symbol = %v, want BTCUSDT", tick.Symbol)
	}
	if tick.Price.String() != "16500.5" {
		t.Errorf("Price = %v, want 16500.5", tick.Price)
	}
	if tick.Volume.String() != "0.123" {
		t.Errorf("Volume = %v, want 0.123", tick.Volume)
	}
}

// TestParseKlineMessage 测试解析 K线 消息
func TestParseKlineMessage(t *testing.T) {
	client := NewWSClient(Config{})

	// Binance 真实消息格式（已关闭的 K线）
	data := []byte(`{
		"e": "kline",
		"E": 1672531200000,
		"s": "BTCUSDT",
		"k": {
			"t": 1672531140000,
			"T": 1672531199999,
			"s": "BTCUSDT",
			"i": "1m",
			"o": "16500.00",
			"h": "16550.00",
			"l": "16490.00",
			"c": "16520.00",
			"v": "123.456",
			"x": true
		}
	}`)

	candle := client.parseKlineMessage(data)
	if candle == nil {
		t.Fatal("parseKlineMessage returned nil")
	}

	if candle.Symbol != "BTCUSDT" {
		t.Errorf("Symbol = %v, want BTCUSDT", candle.Symbol)
	}
	if candle.Interval != "1m" {
		t.Errorf("Interval = %v, want 1m", candle.Interval)
	}
	if candle.Open.String() != "16500" {
		t.Errorf("Open = %v, want 16500", candle.Open)
	}
	if candle.Close.String() != "16520" {
		t.Errorf("Close = %v, want 16520", candle.Close)
	}

	// 测试未关闭的 K线（应返回 nil）
	dataOpen := []byte(`{
		"e": "kline",
		"E": 1672531200000,
		"s": "BTCUSDT",
		"k": {
			"t": 1672531140000,
			"T": 1672531199999,
			"s": "BTCUSDT",
			"i": "1m",
			"o": "16500.00",
			"h": "16550.00",
			"l": "16490.00",
			"c": "16520.00",
			"v": "123.456",
			"x": false
		}
	}`)

	candleOpen := client.parseKlineMessage(dataOpen)
	if candleOpen != nil {
		t.Error("parseKlineMessage should return nil for open kline")
	}
}

// TestWSClient_GetHistoricalKLines 测试拉取历史 K线
func TestWSClient_GetHistoricalKLines(t *testing.T) {
	apiKey := os.Getenv("BINANCE_API_KEY")
	apiSecret := os.Getenv("BINANCE_API_SECRET")

	if apiKey == "" || apiSecret == "" {
		t.Skip("Skipping integration test: BINANCE_API_KEY or BINANCE_API_SECRET not set")
	}

	client := NewWSClient(Config{
		APIKey:    apiKey,
		APISecret: apiSecret,
		Testnet:   true,
	})

	ctx := context.Background()

	// 拉取最近 1 小时的 1m K线
	endTime := time.Now().UnixMilli()
	startTime := endTime - 3600*1000 // 1 hour ago

	candles, err := client.GetHistoricalKLines(ctx, "BTCUSDT", "1m", startTime, endTime)
	if err != nil {
		t.Fatalf("GetHistoricalKLines failed: %v", err)
	}

	if len(candles) == 0 {
		t.Error("No candles returned")
	}

	t.Logf("Got %d candles", len(candles))
	if len(candles) > 0 {
		c := candles[0]
		t.Logf("First candle: %s [%s] O:%s H:%s L:%s C:%s V:%s",
			c.Symbol, c.Interval, c.Open.String(), c.High.String(),
			c.Low.String(), c.Close.String(), c.Volume.String())
	}
}

// TestWSClient_GetLatestPrice 测试获取最新价格
func TestWSClient_GetLatestPrice(t *testing.T) {
	apiKey := os.Getenv("BINANCE_API_KEY")
	apiSecret := os.Getenv("BINANCE_API_SECRET")

	if apiKey == "" || apiSecret == "" {
		t.Skip("Skipping integration test: BINANCE_API_KEY or BINANCE_API_SECRET not set")
	}

	client := NewWSClient(Config{
		APIKey:    apiKey,
		APISecret: apiSecret,
		Testnet:   true,
	})

	ctx := context.Background()

	price, err := client.GetLatestPrice(ctx, "BTCUSDT")
	if err != nil {
		t.Fatalf("GetLatestPrice failed: %v", err)
	}

	if price.IsZero() {
		t.Error("Price is zero")
	}

	t.Logf("Latest BTCUSDT price: %s", price.String())
}
