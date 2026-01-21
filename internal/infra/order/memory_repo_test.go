package order

import (
	"context"
	"testing"
	"time"

	"github.com/iluyuns/alpha-trade/internal/domain/model"
)

func TestMemoryRepo_SaveAndGet(t *testing.T) {
	repo := NewMemoryRepo()
	ctx := context.Background()

	order := &model.Order{
		ClientOrderID: "test-order-1",
		ExchangeID:    "binance-123",
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

	// 根据 ClientOrderID 获取
	loaded, err := repo.GetOrder(ctx, "test-order-1")
	if err != nil {
		t.Fatalf("GetOrder failed: %v", err)
	}

	if loaded.ClientOrderID != "test-order-1" {
		t.Errorf("ClientOrderID = %s, want test-order-1", loaded.ClientOrderID)
	}
	if !loaded.Price.EQ(model.MustMoney("50000")) {
		t.Errorf("Price = %s, want 50000", loaded.Price)
	}

	// 根据 ExchangeID 获取
	loadedByEx, err := repo.GetOrderByExchangeID(ctx, "binance-123")
	if err != nil {
		t.Fatalf("GetOrderByExchangeID failed: %v", err)
	}

	if loadedByEx.ClientOrderID != "test-order-1" {
		t.Errorf("ClientOrderID = %s, want test-order-1", loadedByEx.ClientOrderID)
	}
}

func TestMemoryRepo_UpdateStatus(t *testing.T) {
	repo := NewMemoryRepo()
	ctx := context.Background()

	order := &model.Order{
		ClientOrderID: "test-order-2",
		Symbol:        "ETHUSDT",
		Side:          model.OrderSideSell,
		Type:          model.OrderTypeMarket,
		Quantity:      model.MustMoney("1"),
		Status:        model.OrderStatusPending,
		CreatedAt:     time.Now(),
	}

	_ = repo.SaveOrder(ctx, order)

	// 更新状态
	err := repo.UpdateOrderStatus(ctx, "test-order-2", model.OrderStatusFilled)
	if err != nil {
		t.Fatalf("UpdateOrderStatus failed: %v", err)
	}

	// 验证更新
	loaded, _ := repo.GetOrder(ctx, "test-order-2")
	if loaded.Status != model.OrderStatusFilled {
		t.Errorf("Status = %v, want FILLED", loaded.Status)
	}
}

func TestMemoryRepo_UpdateFilled(t *testing.T) {
	repo := NewMemoryRepo()
	ctx := context.Background()

	order := &model.Order{
		ClientOrderID: "test-order-3",
		Symbol:        "BTCUSDT",
		Side:          model.OrderSideBuy,
		Quantity:      model.MustMoney("1"),
		Filled:        model.MustMoney("0"),
		Status:        model.OrderStatusSubmitted,
		CreatedAt:     time.Now(),
	}

	_ = repo.SaveOrder(ctx, order)

	// 更新成交数量
	err := repo.UpdateFilled(ctx, "test-order-3", model.MustMoney("0.5"))
	if err != nil {
		t.Fatalf("UpdateFilled failed: %v", err)
	}

	// 验证更新
	loaded, _ := repo.GetOrder(ctx, "test-order-3")
	if !loaded.Filled.EQ(model.MustMoney("0.5")) {
		t.Errorf("Filled = %s, want 0.5", loaded.Filled)
	}
}

func TestMemoryRepo_ListActiveOrders(t *testing.T) {
	repo := NewMemoryRepo()
	ctx := context.Background()

	// 创建多个订单
	orders := []*model.Order{
		{
			ClientOrderID: "active-1",
			Symbol:        "BTCUSDT",
			Status:        model.OrderStatusSubmitted,
			CreatedAt:     time.Now(),
		},
		{
			ClientOrderID: "active-2",
			Symbol:        "ETHUSDT",
			Status:        model.OrderStatusPartialFilled,
			CreatedAt:     time.Now(),
		},
		{
			ClientOrderID: "closed-1",
			Symbol:        "BTCUSDT",
			Status:        model.OrderStatusFilled,
			CreatedAt:     time.Now(),
		},
	}

	for _, order := range orders {
		_ = repo.SaveOrder(ctx, order)
	}

	// 列出活跃订单
	activeOrders, err := repo.ListActiveOrders(ctx)
	if err != nil {
		t.Fatalf("ListActiveOrders failed: %v", err)
	}

	if len(activeOrders) != 2 {
		t.Errorf("Active orders count = %d, want 2", len(activeOrders))
	}
}

func TestMemoryRepo_ListOrdersBySymbol(t *testing.T) {
	repo := NewMemoryRepo()
	ctx := context.Background()

	// 创建多个订单
	orders := []*model.Order{
		{
			ClientOrderID: "btc-1",
			Symbol:        "BTCUSDT",
			Status:        model.OrderStatusFilled,
			CreatedAt:     time.Now(),
		},
		{
			ClientOrderID: "btc-2",
			Symbol:        "BTCUSDT",
			Status:        model.OrderStatusCancelled,
			CreatedAt:     time.Now(),
		},
		{
			ClientOrderID: "eth-1",
			Symbol:        "ETHUSDT",
			Status:        model.OrderStatusFilled,
			CreatedAt:     time.Now(),
		},
	}

	for _, order := range orders {
		_ = repo.SaveOrder(ctx, order)
	}

	// 列出 BTCUSDT 订单
	btcOrders, err := repo.ListOrdersBySymbol(ctx, "BTCUSDT", 10)
	if err != nil {
		t.Fatalf("ListOrdersBySymbol failed: %v", err)
	}

	if len(btcOrders) != 2 {
		t.Errorf("BTCUSDT orders count = %d, want 2", len(btcOrders))
	}
}
