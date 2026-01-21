package risk

import (
	"context"
	"testing"

	"github.com/iluyuns/alpha-trade/internal/domain/model"
)

func TestMemoryRiskRepo_SaveAndLoad(t *testing.T) {
	repo := NewMemoryRiskRepo()
	ctx := context.Background()

	// 保存状态
	state := model.NewRiskState("test-account", model.MustMoney("10000"))
	state.DailyPnL = model.MustMoney("-500")
	state.ConsecutiveLosses = 2

	err := repo.SaveState(ctx, state)
	if err != nil {
		t.Fatalf("SaveState failed: %v", err)
	}

	// 加载状态
	loaded, err := repo.LoadState(ctx, "test-account", "")
	if err != nil {
		t.Fatalf("LoadState failed: %v", err)
	}

	if loaded.DailyPnL.String() != "-500" {
		t.Errorf("DailyPnL = %s, want -500", loaded.DailyPnL.String())
	}
	if loaded.ConsecutiveLosses != 2 {
		t.Errorf("ConsecutiveLosses = %d, want 2", loaded.ConsecutiveLosses)
	}
}

func TestMemoryRiskRepo_UpdateEquity(t *testing.T) {
	repo := NewMemoryRiskRepo()
	ctx := context.Background()

	// 初始化
	state := model.NewRiskState("test", model.MustMoney("10000"))
	_ = repo.SaveState(ctx, state)

	// 更新净值
	err := repo.UpdateEquity(ctx, "test", model.MustMoney("12000"))
	if err != nil {
		t.Fatalf("UpdateEquity failed: %v", err)
	}

	// 验证
	loaded, _ := repo.LoadState(ctx, "test", "")
	if loaded.CurrentEquity.String() != "12000" {
		t.Errorf("CurrentEquity = %s, want 12000", loaded.CurrentEquity.String())
	}
	if loaded.PeakEquity.String() != "12000" {
		t.Errorf("PeakEquity = %s, want 12000", loaded.PeakEquity.String())
	}
}

func TestMemoryRiskRepo_RecordTrade(t *testing.T) {
	repo := NewMemoryRiskRepo()
	ctx := context.Background()

	state := model.NewRiskState("test", model.MustMoney("10000"))
	_ = repo.SaveState(ctx, state)

	// 记录亏损
	_ = repo.RecordTrade(ctx, "test", model.MustMoney("-100"))
	_ = repo.RecordTrade(ctx, "test", model.MustMoney("-200"))

	loaded, _ := repo.LoadState(ctx, "test", "")
	if loaded.DailyPnL.String() != "-300" {
		t.Errorf("DailyPnL = %s, want -300", loaded.DailyPnL.String())
	}
	if loaded.ConsecutiveLosses != 2 {
		t.Errorf("ConsecutiveLosses = %d, want 2", loaded.ConsecutiveLosses)
	}
	if loaded.DailyTradeCount != 2 {
		t.Errorf("DailyTradeCount = %d, want 2", loaded.DailyTradeCount)
	}

	// 记录盈利（重置连续亏损）
	_ = repo.RecordTrade(ctx, "test", model.MustMoney("150"))

	loaded, _ = repo.LoadState(ctx, "test", "")
	if loaded.ConsecutiveLosses != 0 {
		t.Errorf("ConsecutiveLosses = %d, want 0 after profit", loaded.ConsecutiveLosses)
	}
}

func TestMemoryRiskRepo_CircuitBreaker(t *testing.T) {
	repo := NewMemoryRiskRepo()
	ctx := context.Background()

	state := model.NewRiskState("test", model.MustMoney("10000"))
	_ = repo.SaveState(ctx, state)

	// 打开熔断器
	err := repo.OpenCircuitBreaker(ctx, "test", 3600) // 1 hour
	if err != nil {
		t.Fatalf("OpenCircuitBreaker failed: %v", err)
	}

	// 检查状态
	isOpen, err := repo.IsCircuitBreakerOpen(ctx, "test")
	if err != nil {
		t.Fatalf("IsCircuitBreakerOpen failed: %v", err)
	}
	if !isOpen {
		t.Error("Expected circuit breaker to be open")
	}

	// 关闭熔断器
	err = repo.CloseCircuitBreaker(ctx, "test")
	if err != nil {
		t.Fatalf("CloseCircuitBreaker failed: %v", err)
	}

	isOpen, _ = repo.IsCircuitBreakerOpen(ctx, "test")
	if isOpen {
		t.Error("Expected circuit breaker to be closed")
	}
}
