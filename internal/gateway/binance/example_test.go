package binance_test

import (
	"context"
	"fmt"
	"time"

	"github.com/iluyuns/alpha-trade/internal/gateway/binance"
)

// ExampleWSClient_SubscribeTicks 演示订阅 Tick 数据
func ExampleWSClient_SubscribeTicks() {
	// 创建 WebSocket 客户端
	client := binance.NewWSClient(binance.Config{
		APIKey:    "your-api-key",
		APISecret: "your-api-secret",
		Testnet:   true,
	})
	defer client.Close()

	// 订阅多个交易对的 Tick 数据
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	symbols := []string{"BTCUSDT", "ETHUSDT"}
	tickCh, err := client.SubscribeTicks(ctx, symbols)
	if err != nil {
		panic(err)
	}

	// 读取 Tick 数据
	for tick := range tickCh {
		fmt.Printf("Tick: %s @ %s\n", tick.Symbol, tick.Price.String())
	}
}

// ExampleWSClient_SubscribeKLines 演示订阅 K线数据
func ExampleWSClient_SubscribeKLines() {
	client := binance.NewWSClient(binance.Config{
		APIKey:    "your-api-key",
		APISecret: "your-api-secret",
		Testnet:   true,
	})
	defer client.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	symbols := []string{"BTCUSDT"}
	klineCh, err := client.SubscribeKLines(ctx, symbols, "1m")
	if err != nil {
		panic(err)
	}

	// 读取 K线数据
	for candle := range klineCh {
		fmt.Printf("Candle: %s [%s] Close: %s\n",
			candle.Symbol, candle.Interval, candle.Close.String())
	}
}

// ExampleWSClient_GetHistoricalKLines 演示拉取历史 K线
func ExampleWSClient_GetHistoricalKLines() {
	client := binance.NewWSClient(binance.Config{
		APIKey:    "your-api-key",
		APISecret: "your-api-secret",
		Testnet:   true,
	})

	ctx := context.Background()

	// 拉取最近 1 小时的 1m K线
	endTime := time.Now().UnixMilli()
	startTime := endTime - 3600*1000

	candles, err := client.GetHistoricalKLines(ctx, "BTCUSDT", "1m", startTime, endTime)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Got %d candles\n", len(candles))
}

// ExampleWSClient_GetLatestPrice 演示获取最新价格
func ExampleWSClient_GetLatestPrice() {
	client := binance.NewWSClient(binance.Config{
		APIKey:    "your-api-key",
		APISecret: "your-api-secret",
		Testnet:   true,
	})

	ctx := context.Background()

	price, err := client.GetLatestPrice(ctx, "BTCUSDT")
	if err != nil {
		panic(err)
	}

	fmt.Printf("Latest BTCUSDT price: %s\n", price.String())
}
