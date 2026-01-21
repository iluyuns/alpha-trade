package binance

import (
	"context"
	"os"
	"testing"

	"github.com/iluyuns/alpha-trade/internal/domain/model"
)

// TestBinanceSpotClient_Integration 集成测试（需要真实 API Key）
func TestBinanceSpotClient_Integration(t *testing.T) {
	apiKey := os.Getenv("BINANCE_API_KEY")
	apiSecret := os.Getenv("BINANCE_API_SECRET")

	if apiKey == "" || apiSecret == "" {
		t.Skip("Skipping integration test: BINANCE_API_KEY or BINANCE_API_SECRET not set")
	}

	client := NewSpotClient(Config{
		APIKey:    apiKey,
		APISecret: apiSecret,
		Testnet:   true, // 使用测试网
	})

	ctx := context.Background()

	t.Run("GetAllBalances", func(t *testing.T) {
		balances, err := client.GetAllBalances(ctx)
		if err != nil {
			t.Fatalf("GetAllBalances failed: %v", err)
		}

		t.Logf("Got %d balances", len(balances))
		for _, bal := range balances {
			t.Logf("  %s: Free=%s, Locked=%s, Total=%s",
				bal.Asset, bal.Free.String(), bal.Locked.String(), bal.Total.String())
		}
	})

	t.Run("GetBalance", func(t *testing.T) {
		balance, err := client.GetBalance(ctx, "USDT")
		if err != nil {
			t.Fatalf("GetBalance failed: %v", err)
		}

		t.Logf("USDT Balance: Free=%s, Locked=%s, Total=%s",
			balance.Free.String(), balance.Locked.String(), balance.Total.String())
	})
}

func TestConvertOrderType(t *testing.T) {
	client := &SpotClient{}

	tests := []struct {
		input model.OrderType
		want  string
	}{
		{model.OrderTypeMarket, "MARKET"},
		{model.OrderTypeLimit, "LIMIT"},
		{model.OrderTypeIOC, "LIMIT"},
		{model.OrderTypeFOK, "LIMIT"},
	}

	for _, tt := range tests {
		t.Run(tt.input.String(), func(t *testing.T) {
			got := client.convertOrderType(tt.input)
			if got != tt.want {
				t.Errorf("convertOrderType() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConvertSide(t *testing.T) {
	client := &SpotClient{}

	tests := []struct {
		input model.OrderSide
		want  string
	}{
		{model.OrderSideBuy, "BUY"},
		{model.OrderSideSell, "SELL"},
	}

	for _, tt := range tests {
		t.Run(tt.input.String(), func(t *testing.T) {
			got := client.convertSide(tt.input)
			if got != tt.want {
				t.Errorf("convertSide() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConvertOrderStatus(t *testing.T) {
	client := &SpotClient{}

	tests := []struct {
		input string
		want  model.OrderStatus
	}{
		{"NEW", model.OrderStatusSubmitted},
		{"PARTIALLY_FILLED", model.OrderStatusPartialFilled},
		{"FILLED", model.OrderStatusFilled},
		{"CANCELED", model.OrderStatusCancelled},
		{"REJECTED", model.OrderStatusRejected},
		{"EXPIRED", model.OrderStatusCancelled},
		{"UNKNOWN", model.OrderStatusPending},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := client.convertOrderStatus(tt.input)
			if got != tt.want {
				t.Errorf("convertOrderStatus() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestExtractSymbolFromOrderID(t *testing.T) {
	client := &SpotClient{}

	tests := []struct {
		name    string
		orderID string
		want    string
	}{
		{"with symbol prefix", "BTCUSDT-uuid-123", "BTCUSDT"},
		{"short symbol", "BTC-uuid-456", "BTC"},
		{"no separator", "BTCUSDT123", ""},
		{"empty", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := client.extractSymbolFromOrderID(tt.orderID)
			if got != tt.want {
				t.Errorf("extractSymbolFromOrderID() = %v, want %v", got, tt.want)
			}
		})
	}
}

// 模拟测试（不需要真实 API）
func TestSpotClient_MockUsage(t *testing.T) {
	// 测试客户端初始化
	client := NewSpotClient(Config{
		APIKey:    "test_key",
		APISecret: "test_secret",
		Testnet:   true,
	})

	if client == nil {
		t.Fatal("NewSpotClient returned nil")
	}

	if client.apiKey != "test_key" {
		t.Errorf("apiKey = %v, want %v", client.apiKey, "test_key")
	}
}

// Benchmark 测试
func BenchmarkConvertOrderType(b *testing.B) {
	client := &SpotClient{}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		client.convertOrderType(model.OrderTypeLimit)
	}
}

func BenchmarkConvertSide(b *testing.B) {
	client := &SpotClient{}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		client.convertSide(model.OrderSideBuy)
	}
}
