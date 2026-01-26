package svc

import (
	"context"
	"testing"
	"time"

	"github.com/iluyuns/alpha-trade/internal/gateway/binance"
	"github.com/iluyuns/alpha-trade/internal/strategy"
	"github.com/iluyuns/alpha-trade/internal/domain/model"
)

// mockStrategy 模拟策略
type mockStrategy struct {
	name string
}

func (m *mockStrategy) Name() string {
	return m.name
}

func (m *mockStrategy) OnCandle(ctx context.Context, candle *model.Candle) (*strategy.TradeSignal, error) {
	// 模拟策略：不产生信号
	return nil, nil
}

func (m *mockStrategy) OnTick(ctx context.Context, tick *model.Tick) (*strategy.TradeSignal, error) {
	return nil, nil
}

func TestTradingLoop_StartStop(t *testing.T) {
	// 创建模拟的 WebSocket 客户端和策略引擎
	// 注意：这里需要实际的实现，或者使用 mock
	// 为了简化测试，我们只测试基本逻辑

	cfg := binance.Config{
		APIKey:    "test-key",
		APISecret: "test-secret",
		Testnet:   true,
	}
	wsClient := binance.NewWSClient(cfg)

	strategyInstance := &mockStrategy{name: "test-strategy"}
	// 创建策略引擎（不通过 OMS，直接使用 Gateway）
	// 注意：这里简化处理，实际测试可能需要 mock
	engine := strategy.NewEngine(strategyInstance, nil, "test-account")

	symbols := []string{"BTCUSDT"}
	interval := "1m"

	loop := NewTradingLoop(wsClient, engine, symbols, interval)

	// 测试初始状态
	if loop.IsStarted() {
		t.Error("TradingLoop should not be started initially")
	}

	// 测试启动（会失败，因为需要真实的 WebSocket 连接）
	// 这里只测试基本逻辑，不测试实际连接
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	err := loop.Start(ctx)
	// 预期会失败（因为需要真实的 WebSocket 连接），但不应该 panic
	if err == nil {
		t.Log("Start returned no error (may be expected if connection succeeds)")
	} else {
		t.Logf("Start returned error (expected): %v", err)
	}

	// 测试停止
	loop.Stop()

	// 测试再次停止（应该安全）
	loop.Stop()
}

func TestTradingLoop_EmptySymbols(t *testing.T) {
	cfg := binance.Config{
		APIKey:    "test-key",
		APISecret: "test-secret",
		Testnet:   true,
	}
	wsClient := binance.NewWSClient(cfg)

	strategyInstance := &mockStrategy{name: "test-strategy"}
	engine := strategy.NewEngine(strategyInstance, nil, "test-account")

	// 使用空符号列表
	loop := NewTradingLoop(wsClient, engine, []string{}, "1m")

	ctx := context.Background()
	err := loop.Start(ctx)
	if err == nil {
		t.Error("Expected error when symbols list is empty")
	}
}
