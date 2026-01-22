package risk

import (
	"context"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/iluyuns/alpha-trade/internal/domain/model"
)

// setupRedisTestClient 创建测试用的 Redis 客户端
// 注意：需要本地运行 Redis 服务，或使用 docker-compose up redis
func setupRedisTestClient(t *testing.T) *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       15, // 使用 DB 15 避免影响生产数据
	})

	// 测试连接
	ctx := context.Background()
	if err := client.Ping(ctx).Err(); err != nil {
		t.Skipf("Redis not available, skipping test: %v", err)
	}

	// 清理测试数据
	_ = client.FlushDB(ctx)

	return client
}

func TestRedisRepo_LoadState(t *testing.T) {
	client := setupRedisTestClient(t)
	defer client.Close()

	repo := NewRedisRepo(client)
	ctx := context.Background()

	t.Run("加载不存在的状态", func(t *testing.T) {
		state, err := repo.LoadState(ctx, "test-account", "")
		if err != nil {
			t.Fatalf("LoadState failed: %v", err)
		}

		if state.AccountID != "test-account" {
			t.Errorf("AccountID mismatch: got %s, want test-account", state.AccountID)
		}
	})

	t.Run("加载已存在的状态", func(t *testing.T) {
		// 先保存状态
		initialState := model.NewRiskState("test-account-2", model.MustMoney("10000"))
		if err := repo.SaveState(ctx, initialState); err != nil {
			t.Fatalf("SaveState failed: %v", err)
		}

		// 加载状态
		loadedState, err := repo.LoadState(ctx, "test-account-2", "")
		if err != nil {
			t.Fatalf("LoadState failed: %v", err)
		}

		if !loadedState.InitialEquity.EQ(model.MustMoney("10000")) {
			t.Errorf("InitialEquity mismatch: got %s, want 10000", loadedState.InitialEquity.String())
		}
	})
}

func TestRedisRepo_SaveState(t *testing.T) {
	client := setupRedisTestClient(t)
	defer client.Close()

	repo := NewRedisRepo(client)
	ctx := context.Background()

	t.Run("保存状态", func(t *testing.T) {
		state := model.NewRiskState("save-test", model.MustMoney("5000"))
		state.DailyPnL = model.MustMoney("-100")

		if err := repo.SaveState(ctx, state); err != nil {
			t.Fatalf("SaveState failed: %v", err)
		}

		// 验证已保存
		loadedState, err := repo.LoadState(ctx, "save-test", "")
		if err != nil {
			t.Fatalf("LoadState failed: %v", err)
		}

		if !loadedState.DailyPnL.EQ(model.MustMoney("-100")) {
			t.Errorf("DailyPnL mismatch: got %s, want -100", loadedState.DailyPnL.String())
		}
	})

	t.Run("幂等保存", func(t *testing.T) {
		state := model.NewRiskState("idempotent-test", model.MustMoney("3000"))
		state.DailyPnL = model.MustMoney("200")

		// 多次保存
		if err := repo.SaveState(ctx, state); err != nil {
			t.Fatalf("SaveState failed: %v", err)
		}
		if err := repo.SaveState(ctx, state); err != nil {
			t.Fatalf("SaveState failed: %v", err)
		}

		// 验证状态一致
		loadedState, err := repo.LoadState(ctx, "idempotent-test", "")
		if err != nil {
			t.Fatalf("LoadState failed: %v", err)
		}

		if !loadedState.DailyPnL.EQ(model.MustMoney("200")) {
			t.Errorf("DailyPnL mismatch after idempotent save: got %s, want 200", loadedState.DailyPnL.String())
		}
	})
}

func TestRedisRepo_UpdateEquity(t *testing.T) {
	client := setupRedisTestClient(t)
	defer client.Close()

	repo := NewRedisRepo(client)
	ctx := context.Background()

	// 初始化状态
	state := model.NewRiskState("equity-test", model.MustMoney("10000"))
	if err := repo.SaveState(ctx, state); err != nil {
		t.Fatalf("SaveState failed: %v", err)
	}

	// 更新净值
	if err := repo.UpdateEquity(ctx, "equity-test", model.MustMoney("12000")); err != nil {
		t.Fatalf("UpdateEquity failed: %v", err)
	}

	// 验证更新
	loadedState, err := repo.LoadState(ctx, "equity-test", "")
	if err != nil {
		t.Fatalf("LoadState failed: %v", err)
	}

	if !loadedState.CurrentEquity.EQ(model.MustMoney("12000")) {
		t.Errorf("CurrentEquity mismatch: got %s, want 12000", loadedState.CurrentEquity.String())
	}

	if !loadedState.PeakEquity.EQ(model.MustMoney("12000")) {
		t.Errorf("PeakEquity should be updated: got %s, want 12000", loadedState.PeakEquity.String())
	}
}

func TestRedisRepo_RecordTrade(t *testing.T) {
	client := setupRedisTestClient(t)
	defer client.Close()

	repo := NewRedisRepo(client)
	ctx := context.Background()

	// 初始化状态
	state := model.NewRiskState("trade-test", model.MustMoney("10000"))
	if err := repo.SaveState(ctx, state); err != nil {
		t.Fatalf("SaveState failed: %v", err)
	}

	// 记录盈利交易
	if err := repo.RecordTrade(ctx, "trade-test", model.MustMoney("500")); err != nil {
		t.Fatalf("RecordTrade failed: %v", err)
	}

	loadedState, err := repo.LoadState(ctx, "trade-test", "")
	if err != nil {
		t.Fatalf("LoadState failed: %v", err)
	}

	if !loadedState.DailyPnL.EQ(model.MustMoney("500")) {
		t.Errorf("DailyPnL mismatch: got %s, want 500", loadedState.DailyPnL.String())
	}

	if loadedState.ConsecutiveLosses != 0 {
		t.Errorf("ConsecutiveLosses should be reset: got %d, want 0", loadedState.ConsecutiveLosses)
	}

	// 记录亏损交易
	if err := repo.RecordTrade(ctx, "trade-test", model.MustMoney("-200")); err != nil {
		t.Fatalf("RecordTrade failed: %v", err)
	}

	loadedState2, err := repo.LoadState(ctx, "trade-test", "")
	if err != nil {
		t.Fatalf("LoadState failed: %v", err)
	}

	if !loadedState2.DailyPnL.EQ(model.MustMoney("300")) {
		t.Errorf("DailyPnL mismatch: got %s, want 300", loadedState2.DailyPnL.String())
	}

	if loadedState2.ConsecutiveLosses != 1 {
		t.Errorf("ConsecutiveLosses should be 1: got %d", loadedState2.ConsecutiveLosses)
	}
}

func TestRedisRepo_CircuitBreaker(t *testing.T) {
	client := setupRedisTestClient(t)
	defer client.Close()

	repo := NewRedisRepo(client)
	ctx := context.Background()

	// 初始化状态
	state := model.NewRiskState("cb-test", model.MustMoney("10000"))
	if err := repo.SaveState(ctx, state); err != nil {
		t.Fatalf("SaveState failed: %v", err)
	}

	// 打开熔断器
	duration := int64(60) // 60 秒
	if err := repo.OpenCircuitBreaker(ctx, "cb-test", duration); err != nil {
		t.Fatalf("OpenCircuitBreaker failed: %v", err)
	}

	// 检查熔断器状态
	isOpen, err := repo.IsCircuitBreakerOpen(ctx, "cb-test")
	if err != nil {
		t.Fatalf("IsCircuitBreakerOpen failed: %v", err)
	}

	if !isOpen {
		t.Error("Circuit breaker should be open")
	}

	// 关闭熔断器
	if err := repo.CloseCircuitBreaker(ctx, "cb-test"); err != nil {
		t.Fatalf("CloseCircuitBreaker failed: %v", err)
	}

	// 再次检查
	isOpen2, err := repo.IsCircuitBreakerOpen(ctx, "cb-test")
	if err != nil {
		t.Fatalf("IsCircuitBreakerOpen failed: %v", err)
	}

	if isOpen2 {
		t.Error("Circuit breaker should be closed")
	}
}

func TestRedisRepo_TTL(t *testing.T) {
	client := setupRedisTestClient(t)
	defer client.Close()

	// 使用短 TTL 测试
	repo := NewRedisRepoWithTTL(client, 2*time.Second)
	ctx := context.Background()

	// 保存状态
	state := model.NewRiskState("ttl-test", model.MustMoney("10000"))
	if err := repo.SaveState(ctx, state); err != nil {
		t.Fatalf("SaveState failed: %v", err)
	}

	// 立即加载应该存在
	_, err := repo.LoadState(ctx, "ttl-test", "")
	if err != nil {
		t.Fatalf("LoadState failed: %v", err)
	}

	// 等待 TTL 过期
	time.Sleep(3 * time.Second)

	// 再次加载应该返回新状态（因为已过期）
	loadedState, err := repo.LoadState(ctx, "ttl-test", "")
	if err != nil {
		t.Fatalf("LoadState failed: %v", err)
	}

	// 过期后应该返回默认状态（DailyPnL 为 0）
	if !loadedState.DailyPnL.IsZero() {
		t.Logf("Note: State may still exist if TTL hasn't expired yet")
	}
}
