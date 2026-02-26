package cache

import (
	"context"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/iluyuns/alpha-trade/internal/domain/model"
	"github.com/iluyuns/alpha-trade/internal/domain/port"
)

// mockMarketData 简单的 mock MarketDataRepo
type mockMarketData struct {
	prices map[string]model.Money
	calls  int
}

func newMockMarketData() *mockMarketData {
	return &mockMarketData{
		prices: map[string]model.Money{
			"BTCUSDT": model.MustMoney("65000.50"),
			"ETHUSDT": model.MustMoney("3200.25"),
		},
	}
}

func (m *mockMarketData) GetLatestPrice(_ context.Context, symbol string) (model.Money, error) {
	m.calls++
	if p, ok := m.prices[symbol]; ok {
		return p, nil
	}
	return model.Zero(), nil
}

func (m *mockMarketData) SubscribeTicks(_ context.Context, _ []string) (<-chan *model.Tick, error) {
	return nil, nil
}
func (m *mockMarketData) SubscribeKLines(_ context.Context, _ []string, _ string) (<-chan *model.Candle, error) {
	return nil, nil
}
func (m *mockMarketData) GetHistoricalKLines(_ context.Context, _ string, _ string, _, _ int64) ([]*model.Candle, error) {
	return nil, nil
}

var _ port.MarketDataRepo = (*mockMarketData)(nil)

func TestPriceCache_GetLatestPrice(t *testing.T) {
	client := redis.NewClient(&redis.Options{Addr: "localhost:6379", DB: 14})
	ctx := context.Background()
	if err := client.Ping(ctx).Err(); err != nil {
		t.Skipf("Redis not available: %v", err)
	}
	_ = client.FlushDB(ctx)
	defer client.Close()

	mock := newMockMarketData()
	cache := NewPriceCache(mock, client, 5*time.Second)

	// 第一次调用：缓存 miss，回源
	price, err := cache.GetLatestPrice(ctx, "BTCUSDT")
	if err != nil {
		t.Fatalf("GetLatestPrice failed: %v", err)
	}
	if !price.EQ(model.MustMoney("65000.50")) {
		t.Errorf("got %s, want 65000.50", price.String())
	}
	if mock.calls != 1 {
		t.Errorf("upstream calls: got %d, want 1", mock.calls)
	}

	// 第二次调用：缓存命中，不回源
	price2, err := cache.GetLatestPrice(ctx, "BTCUSDT")
	if err != nil {
		t.Fatalf("GetLatestPrice (cached) failed: %v", err)
	}
	if !price2.EQ(model.MustMoney("65000.50")) {
		t.Errorf("cached got %s, want 65000.50", price2.String())
	}
	if mock.calls != 1 {
		t.Errorf("upstream calls after cache hit: got %d, want 1", mock.calls)
	}
}

func TestPriceCache_SetPrice(t *testing.T) {
	client := redis.NewClient(&redis.Options{Addr: "localhost:6379", DB: 14})
	ctx := context.Background()
	if err := client.Ping(ctx).Err(); err != nil {
		t.Skipf("Redis not available: %v", err)
	}
	_ = client.FlushDB(ctx)
	defer client.Close()

	mock := newMockMarketData()
	cache := NewPriceCache(mock, client, 5*time.Second)

	// 手动写入价格（模拟 WebSocket 回调）
	_ = cache.SetPrice(ctx, "BNBUSDT", model.MustMoney("580.00"))

	// 读取应命中缓存
	price, err := cache.GetLatestPrice(ctx, "BNBUSDT")
	if err != nil {
		t.Fatalf("GetLatestPrice failed: %v", err)
	}
	if !price.EQ(model.MustMoney("580.00")) {
		t.Errorf("got %s, want 580.00", price.String())
	}
	if mock.calls != 0 {
		t.Errorf("upstream should not be called: got %d calls", mock.calls)
	}
}
