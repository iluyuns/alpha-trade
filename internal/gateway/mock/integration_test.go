package mock

import (
	"context"
	"testing"

	"github.com/iluyuns/alpha-trade/internal/domain/model"
	"github.com/iluyuns/alpha-trade/internal/domain/port"
	"github.com/iluyuns/alpha-trade/internal/infra/risk"
	risklogic "github.com/iluyuns/alpha-trade/internal/logic/risk"
)

// TestIntegration_RiskManager_MockExchange 集成测试
// 验证风控系统与模拟交易所协同工作
func TestIntegration_RiskManager_MockExchange(t *testing.T) {
	ctx := context.Background()

	// 1. 初始化模拟交易所
	exchange := NewSpotExchange(map[string]model.Money{
		"USDT": model.MustMoney("10000"),
		"BTC":  model.MustMoney("0"),
	})
	exchange.SetPrice("BTCUSDT", model.MustMoney("50000"))

	// 2. 初始化风控系统
	riskRepo := risk.NewMemoryRiskRepo()
	riskMgr := risklogic.NewManager(riskRepo, risklogic.RiskConfig{
		MaxSinglePositionPercent: 0.3,  // 单标的最大30%
		MaxTotalExposurePercent:  0.7,  // 总敞口最大70%
		MinCashReservePercent:    0.3,  // 最小现金储备30%
		MaxConsecutiveLosses:     3,    // 最大连续亏损3次
		MaxDailyDrawdown:         0.05, // 最大日回撤5%
	})

	// 初始化风控状态
	state := model.NewRiskState("test-account", model.MustMoney("10000"))
	_ = riskRepo.SaveState(ctx, state)

	// 3. 场景1：正常下单（通过风控）
	t.Run("允许正常订单", func(t *testing.T) {
		orderReq := &risklogic.OrderContext{
			ClientOrderID: "order-1",
			Symbol:        "BTCUSDT",
			MarketType:    model.MarketTypeSpot,
			Side:          model.OrderSideBuy,
			Type:          model.OrderTypeMarket,
			Price:         model.Zero(),
			Quantity:      model.MustMoney("0.04"), // 2000 USD / 10000 = 20%
			CurrentPrice:  model.MustMoney("50000"),
			AccountID:     "test-account",
		}

		// 风控检查
		decision, err := riskMgr.CheckPreTrade(ctx, orderReq)
		if err != nil {
			t.Fatalf("风控检查失败: %v", err)
		}

		if !decision.IsAllowed() {
			t.Fatalf("订单被拒绝: %s", decision.Reason)
		}

		// 下单
		order, err := exchange.PlaceOrder(ctx, &port.SpotPlaceOrderRequest{
			ClientOrderID: orderReq.ClientOrderID,
			Symbol:        orderReq.Symbol,
			Side:          orderReq.Side,
			Type:          orderReq.Type,
			Price:         orderReq.Price,
			Quantity:      orderReq.Quantity,
		})

		if err != nil {
			t.Fatalf("下单失败: %v", err)
		}

		if !order.IsFilled() {
			t.Error("订单未成交")
		}

		// 验证余额变化
		btcBal, _ := exchange.GetBalance(ctx, "BTC")
		if btcBal.Free.LT(model.MustMoney("0.039")) { // 扣除手续费后应该接近0.04
			t.Errorf("BTC余额异常: %s", btcBal.Free.String())
		}
	})

	// 4. 场景2：超限订单（被风控阻止）
	t.Run("阻止超限订单", func(t *testing.T) {
		orderReq := &risklogic.OrderContext{
			ClientOrderID: "order-2",
			Symbol:        "BTCUSDT",
			MarketType:    model.MarketTypeSpot,
			Side:          model.OrderSideBuy,
			Type:          model.OrderTypeMarket,
			Price:         model.Zero(),
			Quantity:      model.MustMoney("0.1"), // 5000 USD / 10000 = 50% > 30% limit
			CurrentPrice:  model.MustMoney("50000"),
			AccountID:     "test-account",
		}

		decision, err := riskMgr.CheckPreTrade(ctx, orderReq)
		if err != nil {
			t.Fatalf("风控检查失败: %v", err)
		}

		if decision.IsAllowed() {
			t.Error("超限订单应该被拒绝")
		}

		if !decision.ShouldReduce() {
			t.Error("预期返回降档建议")
		}

		t.Logf("风控拒绝原因: %s", decision.Reason)
		t.Logf("建议数量: %s", decision.SuggestedQuantity)
	})

	// 5. 场景3：连续亏损触发熔断
	t.Run("连续亏损触发熔断", func(t *testing.T) {
		// 模拟连续亏损
		_ = riskRepo.RecordTrade(ctx, "test-account", model.MustMoney("-100"))
		_ = riskRepo.RecordTrade(ctx, "test-account", model.MustMoney("-150"))
		_ = riskRepo.RecordTrade(ctx, "test-account", model.MustMoney("-200"))

		// 清除缓存（重要：让风控管理器重新加载状态）
		riskMgr.InvalidateCache("test-account", "")

		// 下一笔订单应该被熔断器阻止
		orderReq := &risklogic.OrderContext{
			ClientOrderID: "order-3",
			Symbol:        "BTCUSDT",
			MarketType:    model.MarketTypeSpot,
			Side:          model.OrderSideBuy,
			Type:          model.OrderTypeMarket,
			Quantity:      model.MustMoney("0.01"),
			CurrentPrice:  model.MustMoney("50000"),
			AccountID:     "test-account",
		}

		decision, err := riskMgr.CheckPreTrade(ctx, orderReq)
		if err != nil {
			t.Fatalf("风控检查失败: %v", err)
		}

		if !decision.IsBlocked() {
			t.Error("连续亏损后订单应被熔断器阻止")
		}

		t.Logf("熔断原因: %s", decision.Reason)
	})
}
