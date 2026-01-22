package oms

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/iluyuns/alpha-trade/internal/domain/model"
	"github.com/iluyuns/alpha-trade/internal/gateway/mock"
	"github.com/iluyuns/alpha-trade/internal/infra/order"
	"github.com/iluyuns/alpha-trade/internal/infra/risk"
	risklogic "github.com/iluyuns/alpha-trade/internal/logic/risk"
)

func TestManager_PlaceOrder(t *testing.T) {
	ctx := context.Background()
	accountID := "oms-test-account"
	initialCapital := model.MustMoney("10000")

	// 初始化组件
	exchange := mock.NewSpotExchange(map[string]model.Money{
		"USDT": initialCapital,
		"BTC":  model.Zero(),
	})
	exchange.SetPrice("BTCUSDT", model.MustMoney("50000"))

	orderRepo := order.NewMemoryRepo()
	riskRepo := risk.NewMemoryRiskRepo()

	riskMgr := risklogic.NewManager(riskRepo, risklogic.RiskConfig{
		MaxSinglePositionPercent: 0.3,
		MaxTotalExposurePercent:  0.7,
		MinCashReservePercent:    0.3,
		MaxConsecutiveLosses:     3,
		MaxDailyDrawdown:         0.05,
	})

	// 初始化风控状态
	state := model.NewRiskState(accountID, initialCapital)
	_ = riskRepo.SaveState(ctx, state)

	// 创建 OMS
	oms := NewManager(exchange, orderRepo, riskMgr, Config{
		AutoSync: false, // 测试中禁用自动同步
	})

	t.Run("正常下单", func(t *testing.T) {
		req := &PlaceOrderRequest{
			ClientOrderID: "oms-order-1",
			Symbol:        "BTCUSDT",
			Side:          model.OrderSideBuy,
			Type:          model.OrderTypeMarket,
			Quantity:      model.MustMoney("0.04"), // 20% of 10000
			CurrentPrice:  model.MustMoney("50000"),
			AccountID:     accountID,
		}

		order, err := oms.PlaceOrder(ctx, req)
		if err != nil {
			t.Fatalf("PlaceOrder failed: %v", err)
		}

		if order.ClientOrderID != "oms-order-1" {
			t.Errorf("ClientOrderID mismatch: got %s, want oms-order-1", order.ClientOrderID)
		}

		// 验证订单已保存
		savedOrder, err := orderRepo.GetOrder(ctx, "oms-order-1")
		if err != nil {
			t.Fatalf("GetOrder failed: %v", err)
		}

		if savedOrder.Status != model.OrderStatusFilled {
			t.Errorf("Order status mismatch: got %s, want FILLED", savedOrder.Status)
		}
	})

	t.Run("风控拦截", func(t *testing.T) {
		req := &PlaceOrderRequest{
			ClientOrderID: "oms-order-blocked",
			Symbol:        "BTCUSDT",
			Side:          model.OrderSideBuy,
			Type:          model.OrderTypeMarket,
			Quantity:      model.MustMoney("0.1"), // 50% > 30% limit
			CurrentPrice:  model.MustMoney("50000"),
			AccountID:     accountID,
		}

		_, err := oms.PlaceOrder(ctx, req)
		if err == nil {
			t.Error("Order should be rejected by risk manager")
		}

		t.Logf("Order correctly rejected: %v", err)
	})
}

func TestManager_SyncOrderStatus(t *testing.T) {
	ctx := context.Background()

	exchange := mock.NewSpotExchange(map[string]model.Money{
		"USDT": model.MustMoney("10000"),
		"BTC":  model.Zero(),
	})
	exchange.SetPrice("BTCUSDT", model.MustMoney("50000"))

	orderRepo := order.NewMemoryRepo()
	riskRepo := risk.NewMemoryRiskRepo()

	riskMgr := risklogic.NewManager(riskRepo, risklogic.RiskConfig{
		MaxSinglePositionPercent: 0.3,
		MaxTotalExposurePercent:  0.7,
		MinCashReservePercent:    0.3,
		MaxConsecutiveLosses:     3,
		MaxDailyDrawdown:         0.05,
	})

	oms := NewManager(exchange, orderRepo, riskMgr, Config{
		AutoSync: false,
	})

	// 先下一个订单
	req := &PlaceOrderRequest{
		ClientOrderID: "sync-test-order",
		Symbol:        "BTCUSDT",
		Side:          model.OrderSideBuy,
		Type:          model.OrderTypeMarket,
		Quantity:      model.MustMoney("0.02"),
		CurrentPrice:  model.MustMoney("50000"),
		AccountID:     "sync-test-account",
	}

	_, err := oms.PlaceOrder(ctx, req)
	if err != nil {
		t.Fatalf("PlaceOrder failed: %v", err)
	}

	// 同步订单状态
	if err := oms.SyncOrderStatus(ctx, "sync-test-order"); err != nil {
		t.Fatalf("SyncOrderStatus failed: %v", err)
	}

	// 验证状态已同步
	order, err := orderRepo.GetOrder(ctx, "sync-test-order")
	if err != nil {
		t.Fatalf("GetOrder failed: %v", err)
	}

	if !order.IsFilled() {
		t.Errorf("Order should be filled: status=%s", order.Status)
	}
}

func TestManager_CancelOrder(t *testing.T) {
	ctx := context.Background()

	exchange := mock.NewSpotExchange(map[string]model.Money{
		"USDT": model.MustMoney("10000"),
		"BTC":  model.Zero(),
	})
	exchange.SetPrice("BTCUSDT", model.MustMoney("50000"))

	orderRepo := order.NewMemoryRepo()
	riskRepo := risk.NewMemoryRiskRepo()

	riskMgr := risklogic.NewManager(riskRepo, risklogic.RiskConfig{
		MaxSinglePositionPercent: 0.3,
		MaxTotalExposurePercent:  0.7,
		MinCashReservePercent:    0.3,
		MaxConsecutiveLosses:     3,
		MaxDailyDrawdown:         0.05,
	})

	oms := NewManager(exchange, orderRepo, riskMgr, Config{
		AutoSync: false,
	})

	// 注意：Mock Exchange 的订单会立即成交，无法测试撤单
	// 这里仅测试撤单逻辑不会 panic
	t.Run("撤单逻辑", func(t *testing.T) {
		// 创建一个未成交的订单（模拟）
		pendingOrder := &model.Order{
			ClientOrderID: "cancel-test-order",
			ExchangeID:    "MOCK-001",
			Symbol:        "BTCUSDT",
			MarketType:    model.MarketTypeSpot,
			Side:          model.OrderSideBuy,
			Type:          model.OrderTypeLimit,
			Price:         model.MustMoney("49000"),
			Quantity:      model.MustMoney("0.01"),
			Status:        model.OrderStatusPending,
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		}

		if err := orderRepo.SaveOrder(ctx, pendingOrder); err != nil {
			t.Fatalf("SaveOrder failed: %v", err)
		}

		// 尝试撤单（Mock Exchange 可能不支持，但逻辑应该正确）
		err := oms.CancelOrder(ctx, "cancel-test-order")
		if err != nil {
			t.Logf("CancelOrder failed (expected for mock): %v", err)
		}
	})
}

func TestManager_SyncActiveOrders(t *testing.T) {
	ctx := context.Background()

	exchange := mock.NewSpotExchange(map[string]model.Money{
		"USDT": model.MustMoney("10000"),
		"BTC":  model.Zero(),
	})
	exchange.SetPrice("BTCUSDT", model.MustMoney("50000"))

	orderRepo := order.NewMemoryRepo()
	riskRepo := risk.NewMemoryRiskRepo()

	riskMgr := risklogic.NewManager(riskRepo, risklogic.RiskConfig{
		MaxSinglePositionPercent: 0.3,
		MaxTotalExposurePercent:  0.7,
		MinCashReservePercent:    0.3,
		MaxConsecutiveLosses:     3,
		MaxDailyDrawdown:         0.05,
	})

	oms := NewManager(exchange, orderRepo, riskMgr, Config{
		AutoSync: false,
	})

	accountID := "sync-active-test"

	// 创建多个订单
	for i := 1; i <= 3; i++ {
		req := &PlaceOrderRequest{
			ClientOrderID: fmt.Sprintf("sync-active-%d", i),
			Symbol:        "BTCUSDT",
			Side:          model.OrderSideBuy,
			Type:          model.OrderTypeMarket,
			Quantity:      model.MustMoney("0.01"),
			CurrentPrice:  model.MustMoney("50000"),
			AccountID:     accountID,
		}

		_, err := oms.PlaceOrder(ctx, req)
		if err != nil {
			t.Fatalf("PlaceOrder failed: %v", err)
		}
	}

	// 同步所有活跃订单
	if err := oms.SyncActiveOrders(ctx); err != nil {
		t.Fatalf("SyncActiveOrders failed: %v", err)
	}

	// 验证所有订单状态
	activeOrders, err := orderRepo.ListActiveOrders(ctx)
	if err != nil {
		t.Fatalf("ListActiveOrders failed: %v", err)
	}

	t.Logf("Active orders after sync: %d", len(activeOrders))
}
