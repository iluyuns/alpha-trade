package risk

import (
	"context"
	"testing"

	"github.com/iluyuns/alpha-trade/internal/domain/model"
)

func TestPositionLimit_SinglePosition(t *testing.T) {
	tests := []struct {
		name         string
		equity       string
		existingPos  string
		orderSize    string
		price        string
		maxPercent   float64
		wantDecision Decision
	}{
		{
			name:         "under limit",
			equity:       "10000",
			existingPos:  "0",
			orderSize:    "0.05",
			price:        "50000",
			maxPercent:   0.3,
			wantDecision: Allow,
		},
		{
			name:         "at limit",
			equity:       "10000",
			existingPos:  "0",
			orderSize:    "0.06",
			price:        "50000",
			maxPercent:   0.3,
			wantDecision: Allow,
		},
		{
			name:         "over limit - should reduce",
			equity:       "10000",
			existingPos:  "0",
			orderSize:    "0.1",
			price:        "50000",
			maxPercent:   0.3,
			wantDecision: Reduce,
		},
		{
			name:         "existing position at limit",
			equity:       "10000",
			existingPos:  "3000",
			orderSize:    "0.01",
			price:        "50000",
			maxPercent:   0.3,
			wantDecision: Block,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mgr := NewManager(&mockRiskRepo{}, RiskConfig{
				MaxSinglePositionPercent: tt.maxPercent,
			})

			state := model.NewRiskState("test", model.MustMoney(tt.equity))
			state.PositionMap["BTCUSDT"] = model.MustMoney(tt.existingPos)

			req := &OrderContext{
				Symbol:       "BTCUSDT",
				MarketType:   model.MarketTypeSpot,
				Quantity:     model.MustMoney(tt.orderSize),
				Price:        model.MustMoney(tt.price),
				CurrentPrice: model.MustMoney(tt.price),
			}

			decision := mgr.CheckPositionLimit(context.Background(), req, state)

			if decision.Decision != tt.wantDecision {
				t.Errorf("got %v, want %v (reason: %s)", decision.Decision, tt.wantDecision, decision.Reason)
			}
		})
	}
}

func TestPositionLimit_TotalExposure(t *testing.T) {
	mgr := NewManager(&mockRiskRepo{}, RiskConfig{
		MaxTotalExposurePercent: 0.7,
	})

	state := model.NewRiskState("test", model.MustMoney("10000"))
	state.TotalExposure = model.MustMoney("6000") // 60% already

	// New order: 0.05 BTC * 50000 = 2500 USD (25%)
	// Total would be 85% > 70% limit
	req := &OrderContext{
		Symbol:       "BTCUSDT",
		MarketType:   model.MarketTypeSpot,
		Quantity:     model.MustMoney("0.05"),
		Price:        model.MustMoney("50000"),
		CurrentPrice: model.MustMoney("50000"),
	}

	decision := mgr.CheckPositionLimit(context.Background(), req, state)

	if !decision.IsBlocked() {
		t.Errorf("expected blocked, got %v", decision.Decision)
	}
}

func TestPositionLimit_Leverage(t *testing.T) {
	tests := []struct {
		name         string
		leverage     int
		maxLeverage  int
		wantDecision Decision
	}{
		{"under limit", 2, 5, Allow},
		{"at limit", 5, 5, Allow},
		{"over limit", 10, 5, Reduce},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mgr := NewManager(&mockRiskRepo{}, RiskConfig{
				MaxLeverage: tt.maxLeverage,
			})

			state := model.NewRiskState("test", model.MustMoney("10000"))

			req := &OrderContext{
				Symbol:       "BTCUSDT",
				MarketType:   model.MarketTypeFuture,
				Quantity:     model.MustMoney("0.1"),
				Price:        model.MustMoney("50000"),
				CurrentPrice: model.MustMoney("50000"),
				Leverage:     tt.leverage,
			}

			decision := mgr.CheckPositionLimit(context.Background(), req, state)

			if decision.Decision != tt.wantDecision {
				t.Errorf("got %v, want %v", decision.Decision, tt.wantDecision)
			}

			if decision.ShouldReduce() && decision.SuggestedLeverage != tt.maxLeverage {
				t.Errorf("suggested leverage = %d, want %d", decision.SuggestedLeverage, tt.maxLeverage)
			}
		})
	}
}

func TestPositionLimit_LargeOrderForceLeverage(t *testing.T) {
	mgr := NewManager(&mockRiskRepo{}, RiskConfig{
		ForceLeverageOne:    true,
		LargeOrderThreshold: 0.1, // 10% of equity (based on margin, not notional)
	})

	state := model.NewRiskState("test", model.MustMoney("10000"))

	// Order: 0.12 BTC * 50000 = 6000 USD notional
	// With 5x leverage, margin = 6000 / 5 = 1200 USD (12% of equity > 10% threshold)
	req := &OrderContext{
		Symbol:       "BTCUSDT",
		MarketType:   model.MarketTypeFuture,
		Quantity:     model.MustMoney("0.12"),
		Price:        model.MustMoney("50000"),
		CurrentPrice: model.MustMoney("50000"),
		Leverage:     5,
	}

	decision := mgr.CheckPositionLimit(context.Background(), req, state)

	if !decision.ShouldReduce() {
		t.Errorf("expected reduce for large order with leverage, got %v", decision.Decision)
	}

	if decision.SuggestedLeverage != 1 {
		t.Errorf("suggested leverage = %d, want 1", decision.SuggestedLeverage)
	}
}
