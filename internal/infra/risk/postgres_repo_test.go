package risk

import (
	"context"
	"database/sql"
	"os"
	"testing"
	"time"

	_ "github.com/lib/pq"
	"github.com/iluyuns/alpha-trade/internal/domain/model"
)

// TestPostgresRepo_Integration 集成测试（需要真实数据库）
func TestPostgresRepo_Integration(t *testing.T) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("Skipping integration test: DATABASE_URL not set")
	}

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// 确保表存在（运行 migration 后）
	repo := NewPostgresRepo(db)
	ctx := context.Background()
	accountID := "test-account-" + time.Now().Format("20060102150405")

	t.Run("SaveAndLoadState", func(t *testing.T) {
		// 创建新状态
		initialEquity := model.MustMoney("10000")
		state := model.NewRiskState(accountID, initialEquity)
		state.CurrentEquity = model.MustMoney("10500")
		state.PeakEquity = model.MustMoney("11000")
		state.DailyPnL = model.MustMoney("500")
		state.ConsecutiveLosses = 2

		// 保存
		err := repo.SaveState(ctx, state)
		if err != nil {
			t.Fatalf("SaveState failed: %v", err)
		}

		// 加载
		loaded, err := repo.LoadState(ctx, accountID, "")
		if err != nil {
			t.Fatalf("LoadState failed: %v", err)
		}

		// 验证
		if !loaded.InitialEquity.EQ(initialEquity) {
			t.Errorf("InitialEquity = %v, want %v", loaded.InitialEquity, initialEquity)
		}
		if !loaded.CurrentEquity.EQ(model.MustMoney("10500")) {
			t.Errorf("CurrentEquity = %v, want 10500", loaded.CurrentEquity)
		}
		if loaded.ConsecutiveLosses != 2 {
			t.Errorf("ConsecutiveLosses = %d, want 2", loaded.ConsecutiveLosses)
		}
	})

	t.Run("UpdateEquity", func(t *testing.T) {
		newEquity := model.MustMoney("12000")
		err := repo.UpdateEquity(ctx, accountID, newEquity)
		if err != nil {
			t.Fatalf("UpdateEquity failed: %v", err)
		}

		// 验证更新
		loaded, err := repo.LoadState(ctx, accountID, "")
		if err != nil {
			t.Fatalf("LoadState failed: %v", err)
		}

		if !loaded.CurrentEquity.EQ(newEquity) {
			t.Errorf("CurrentEquity = %v, want %v", loaded.CurrentEquity, newEquity)
		}
		// 峰值应该更新
		if !loaded.PeakEquity.EQ(newEquity) {
			t.Errorf("PeakEquity = %v, want %v", loaded.PeakEquity, newEquity)
		}
	})

	t.Run("RecordTrade", func(t *testing.T) {
		// 记录盈利交易
		err := repo.RecordTrade(ctx, accountID, model.MustMoney("100"))
		if err != nil {
			t.Fatalf("RecordTrade (profit) failed: %v", err)
		}

		// 记录亏损交易
		err = repo.RecordTrade(ctx, accountID, model.MustMoney("-50"))
		if err != nil {
			t.Fatalf("RecordTrade (loss) failed: %v", err)
		}

		// 验证
		loaded, _ := repo.LoadState(ctx, accountID, "")
		if !loaded.DailyPnL.EQ(model.MustMoney("550")) { // 500 + 100 - 50
			t.Errorf("DailyPnL = %v, want 550", loaded.DailyPnL)
		}
	})

	t.Run("CircuitBreaker", func(t *testing.T) {
		// 打开熔断器
		err := repo.OpenCircuitBreaker(ctx, accountID, 3600) // 1小时
		if err != nil {
			t.Fatalf("OpenCircuitBreaker failed: %v", err)
		}

		// 检查状态
		isOpen, err := repo.IsCircuitBreakerOpen(ctx, accountID)
		if err != nil {
			t.Fatalf("IsCircuitBreakerOpen failed: %v", err)
		}
		if !isOpen {
			t.Error("Circuit breaker should be open")
		}

		// 关闭熔断器
		err = repo.CloseCircuitBreaker(ctx, accountID)
		if err != nil {
			t.Fatalf("CloseCircuitBreaker failed: %v", err)
		}

		// 验证已关闭
		isOpen, err = repo.IsCircuitBreakerOpen(ctx, accountID)
		if err != nil {
			t.Fatalf("IsCircuitBreakerOpen failed: %v", err)
		}
		if isOpen {
			t.Error("Circuit breaker should be closed")
		}
	})

	// 清理测试数据
	_, _ = db.ExecContext(ctx, "DELETE FROM risk_states WHERE account_id = $1", accountID)
}

// TestPostgresRepo_Unit 单元测试（不需要数据库）
func TestPostgresRepo_Unit(t *testing.T) {
	// 测试仓储初始化
	repo := NewPostgresRepo(nil)
	if repo == nil {
		t.Fatal("NewPostgresRepo returned nil")
	}
	if repo.db != nil {
		t.Error("Expected db to be nil")
	}
}
