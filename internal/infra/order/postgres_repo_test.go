package order

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

	repo := NewPostgresRepo(db)
	ctx := context.Background()
	clientOrderID := "test-order-" + time.Now().Format("20060102150405")

	t.Run("SaveAndGetOrder", func(t *testing.T) {
		order := &model.Order{
			ClientOrderID: clientOrderID,
			ExchangeID:    "binance-test-123",
			Symbol:        "BTCUSDT",
			Side:          model.OrderSideBuy,
			Type:          model.OrderTypeLimit,
			Price:         model.MustMoney("50000"),
			Quantity:      model.MustMoney("0.1"),
			Filled:        model.MustMoney("0"),
			Status:        model.OrderStatusPending,
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		}

		// 保存订单
		err := repo.SaveOrder(ctx, order)
		if err != nil {
			t.Fatalf("SaveOrder failed: %v", err)
		}

		// 获取订单
		loaded, err := repo.GetOrder(ctx, clientOrderID)
		if err != nil {
			t.Fatalf("GetOrder failed: %v", err)
		}

		if loaded.ClientOrderID != clientOrderID {
			t.Errorf("ClientOrderID = %s, want %s", loaded.ClientOrderID, clientOrderID)
		}
		if !loaded.Price.EQ(model.MustMoney("50000")) {
			t.Errorf("Price = %s, want 50000", loaded.Price)
		}
		if loaded.Status != model.OrderStatusPending {
			t.Errorf("Status = %v, want PENDING", loaded.Status)
		}
	})

	t.Run("GetOrderByExchangeID", func(t *testing.T) {
		loaded, err := repo.GetOrderByExchangeID(ctx, "binance-test-123")
		if err != nil {
			t.Fatalf("GetOrderByExchangeID failed: %v", err)
		}

		if loaded.ClientOrderID != clientOrderID {
			t.Errorf("ClientOrderID = %s, want %s", loaded.ClientOrderID, clientOrderID)
		}
	})

	t.Run("UpdateOrderStatus", func(t *testing.T) {
		err := repo.UpdateOrderStatus(ctx, clientOrderID, model.OrderStatusFilled)
		if err != nil {
			t.Fatalf("UpdateOrderStatus failed: %v", err)
		}

		// 验证更新
		loaded, _ := repo.GetOrder(ctx, clientOrderID)
		if loaded.Status != model.OrderStatusFilled {
			t.Errorf("Status = %v, want FILLED", loaded.Status)
		}
	})

	t.Run("UpdateFilled", func(t *testing.T) {
		err := repo.UpdateFilled(ctx, clientOrderID, model.MustMoney("0.05"))
		if err != nil {
			t.Fatalf("UpdateFilled failed: %v", err)
		}

		// 验证更新
		loaded, _ := repo.GetOrder(ctx, clientOrderID)
		if !loaded.Filled.EQ(model.MustMoney("0.05")) {
			t.Errorf("Filled = %s, want 0.05", loaded.Filled)
		}
	})

	t.Run("ListActiveOrders", func(t *testing.T) {
		// 创建一个活跃订单
		activeOrder := &model.Order{
			ClientOrderID: "active-" + time.Now().Format("20060102150405"),
			Symbol:        "ETHUSDT",
			Side:          model.OrderSideSell,
			Type:          model.OrderTypeMarket,
			Quantity:      model.MustMoney("1"),
			Status:        model.OrderStatusSubmitted,
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		}
		_ = repo.SaveOrder(ctx, activeOrder)

		// 列出活跃订单
		orders, err := repo.ListActiveOrders(ctx)
		if err != nil {
			t.Fatalf("ListActiveOrders failed: %v", err)
		}

		if len(orders) == 0 {
			t.Error("Expected at least one active order")
		}
	})

	t.Run("ListOrdersBySymbol", func(t *testing.T) {
		orders, err := repo.ListOrdersBySymbol(ctx, "BTCUSDT", 10)
		if err != nil {
			t.Fatalf("ListOrdersBySymbol failed: %v", err)
		}

		if len(orders) == 0 {
			t.Error("Expected at least one BTCUSDT order")
		}
	})

	// 清理测试数据
	_, _ = db.ExecContext(ctx, "DELETE FROM orders WHERE client_oid LIKE 'test-order-%' OR client_oid LIKE 'active-%'")
}

// TestPostgresRepo_Unit 单元测试（不需要数据库）
func TestPostgresRepo_Unit(t *testing.T) {
	repo := NewPostgresRepo(nil)
	if repo == nil {
		t.Fatal("NewPostgresRepo returned nil")
	}
	if repo.db != nil {
		t.Error("Expected db to be nil")
	}
}
