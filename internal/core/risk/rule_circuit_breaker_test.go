package risk

import (
	"context"
	"testing"
	"time"

	"github.com/iluyuns/alpha-trade/internal/domain/model"
)

func TestCircuitBreaker_ConsecutiveLosses(t *testing.T) {
	tests := []struct {
		name              string
		consecutiveLosses int
		maxAllowed        int
		wantBlocked       bool
	}{
		{"under limit", 2, 3, false},
		{"at limit", 3, 3, true},
		{"over limit", 5, 3, true},
		{"zero config (disabled)", 10, 0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mgr := NewManager(&mockRiskRepo{}, RiskConfig{
				MaxConsecutiveLosses: tt.maxAllowed,
			})

			state := model.NewRiskState("test", model.MustMoney("10000"))
			state.ConsecutiveLosses = tt.consecutiveLosses

			req := &OrderContext{
				Symbol:   "BTCUSDT",
				Quantity: model.MustMoney("0.1"),
			}

			decision := mgr.CheckCircuitBreaker(context.Background(), req, state)

			if tt.wantBlocked && !decision.IsBlocked() {
				t.Errorf("expected blocked, got %v", decision.Decision)
			}
			if !tt.wantBlocked && decision.IsBlocked() {
				t.Errorf("expected allowed, got blocked: %s", decision.Reason)
			}
		})
	}
}

func TestCircuitBreaker_DailyDrawdown(t *testing.T) {
	tests := []struct {
		name        string
		equity      string
		dailyPnL    string
		maxDD       float64
		wantBlocked bool
	}{
		{"no loss", "10000", "0", 0.05, false},
		{"small loss", "10000", "-100", 0.05, false},
		{"at limit", "10000", "-500.01", 0.05, true},
		{"over limit", "10000", "-1000", 0.05, true},
		{"zero config (disabled)", "10000", "-1000", 0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mgr := NewManager(&mockRiskRepo{}, RiskConfig{
				MaxDailyDrawdown: tt.maxDD,
			})

			state := model.NewRiskState("test", model.MustMoney(tt.equity))
			state.DailyPnL = model.MustMoney(tt.dailyPnL)

			req := &OrderContext{
				Symbol:   "BTCUSDT",
				Quantity: model.MustMoney("0.1"),
			}

			decision := mgr.CheckCircuitBreaker(context.Background(), req, state)

			if tt.wantBlocked && !decision.IsBlocked() {
				t.Errorf("expected blocked, got %v", decision.Decision)
			}
			if !tt.wantBlocked && decision.IsBlocked() {
				t.Errorf("expected allowed, got blocked: %s", decision.Reason)
			}
		})
	}
}

func TestCircuitBreaker_AlreadyOpen(t *testing.T) {
	mgr := NewManager(&mockRiskRepo{}, RiskConfig{})

	state := model.NewRiskState("test", model.MustMoney("10000"))
	state.CircuitBreakerOpen = true
	state.CircuitBreakerUntil = time.Now().Add(1 * time.Hour).Unix()

	req := &OrderContext{
		Symbol:   "BTCUSDT",
		Quantity: model.MustMoney("0.1"),
	}

	decision := mgr.CheckCircuitBreaker(context.Background(), req, state)

	if !decision.IsBlocked() {
		t.Error("expected blocked when circuit breaker is open")
	}
}

// Mock RiskRepo for testing
type mockRiskRepo struct{}

func (m *mockRiskRepo) LoadState(ctx context.Context, accountID, symbol string) (*model.RiskState, error) {
	return model.NewRiskState(accountID, model.MustMoney("10000")), nil
}

func (m *mockRiskRepo) SaveState(ctx context.Context, state *model.RiskState) error {
	return nil
}

func (m *mockRiskRepo) UpdateEquity(ctx context.Context, accountID string, newEquity model.Money) error {
	return nil
}

func (m *mockRiskRepo) RecordTrade(ctx context.Context, accountID string, pnl model.Money) error {
	return nil
}

func (m *mockRiskRepo) OpenCircuitBreaker(ctx context.Context, accountID string, duration int64) error {
	return nil
}

func (m *mockRiskRepo) CloseCircuitBreaker(ctx context.Context, accountID string) error {
	return nil
}

func (m *mockRiskRepo) IsCircuitBreakerOpen(ctx context.Context, accountID string) (bool, error) {
	return false, nil
}
