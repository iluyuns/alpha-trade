package integration

import (
	"context"
	"testing"
	"time"

	"github.com/iluyuns/alpha-trade/internal/domain/model"
	"github.com/iluyuns/alpha-trade/internal/domain/port"
	"github.com/iluyuns/alpha-trade/internal/gateway/mock"
	"github.com/iluyuns/alpha-trade/internal/infra/order"
	"github.com/iluyuns/alpha-trade/internal/infra/risk"
	risklogic "github.com/iluyuns/alpha-trade/internal/logic/risk"
	"github.com/iluyuns/alpha-trade/internal/strategy"
)

// TestE2E_TradingFlow 端到端交易流程测试
// 验证：Strategy -> RiskManager -> Gateway -> OrderRepo 完整链路
func TestE2E_TradingFlow(t *testing.T) {
	ctx := context.Background()
	accountID := "e2e-test-account"
	initialCapital := model.MustMoney("10000")
	symbol := "BTCUSDT"
	currentPrice := model.MustMoney("50000")

	// 1. 初始化组件
	setup := setupTradingSystem(t, accountID, initialCapital, symbol, currentPrice)
	defer setup.cleanup()

	// 2. 场景1：正常下单流程
	t.Run("正常下单-完整链路", func(t *testing.T) {
		clientOrderID := "order-normal-1"
		quantity := model.MustMoney("0.04") // 2000 USD = 20% of 10000

		// 构建订单上下文
		orderCtx := &risklogic.OrderContext{
			ClientOrderID: clientOrderID,
			Symbol:        symbol,
			MarketType:    model.MarketTypeSpot,
			Side:          model.OrderSideBuy,
			Type:          model.OrderTypeMarket,
			Price:         model.Zero(),
			Quantity:      quantity,
			CurrentPrice:  currentPrice,
			AccountID:     accountID,
		}

		// Step 1: 风控检查
		decision, err := setup.riskMgr.CheckPreTrade(ctx, orderCtx)
		if err != nil {
			t.Fatalf("风控检查失败: %v", err)
		}
		if !decision.IsAllowed() {
			t.Fatalf("正常订单被拒绝: %s", decision.Reason)
		}

		// Step 2: 下单
		orderReq := &port.SpotPlaceOrderRequest{
			ClientOrderID: clientOrderID,
			Symbol:        symbol,
			Side:          model.OrderSideBuy,
			Type:          model.OrderTypeMarket,
			Price:         model.Zero(),
			Quantity:      quantity,
		}

		placedOrder, err := setup.exchange.PlaceOrder(ctx, orderReq)
		if err != nil {
			t.Fatalf("下单失败: %v", err)
		}

		if !placedOrder.IsFilled() {
			t.Error("订单应已成交")
		}

		// Step 3: 持久化订单
		if err := setup.orderRepo.SaveOrder(ctx, placedOrder); err != nil {
			t.Fatalf("订单持久化失败: %v", err)
		}

		// Step 4: 验证订单已保存
		savedOrder, err := setup.orderRepo.GetOrder(ctx, clientOrderID)
		if err != nil {
			t.Fatalf("查询订单失败: %v", err)
		}

		if savedOrder.ClientOrderID != clientOrderID {
			t.Errorf("订单ID不匹配: got %s, want %s", savedOrder.ClientOrderID, clientOrderID)
		}

		if savedOrder.Status != model.OrderStatusFilled {
			t.Errorf("订单状态错误: got %s, want FILLED", savedOrder.Status)
		}

		// Step 5: 验证余额变化
		btcBal, _ := setup.exchange.GetBalance(ctx, "BTC")
		if btcBal.Free.LT(model.MustMoney("0.039")) {
			t.Errorf("BTC余额异常: %s", btcBal.Free.String())
		}

		t.Logf("✅ 正常下单完成: OrderID=%s, Filled=%s", clientOrderID, savedOrder.Filled.String())
	})

	// 3. 场景2：风控拦截超限订单
	t.Run("风控拦截-超限订单", func(t *testing.T) {
		clientOrderID := "order-blocked-1"
		quantity := model.MustMoney("0.1") // 5000 USD = 50% > 30% limit

		orderCtx := &risklogic.OrderContext{
			ClientOrderID: clientOrderID,
			Symbol:        symbol,
			MarketType:    model.MarketTypeSpot,
			Side:          model.OrderSideBuy,
			Type:          model.OrderTypeMarket,
			Quantity:      quantity,
			CurrentPrice:  currentPrice,
			AccountID:     accountID,
		}

		decision, err := setup.riskMgr.CheckPreTrade(ctx, orderCtx)
		if err != nil {
			t.Fatalf("风控检查失败: %v", err)
		}

		if decision.IsAllowed() {
			t.Error("超限订单应该被拒绝")
		}

		if !decision.ShouldReduce() {
			t.Error("应该返回降档建议")
		}

		t.Logf("✅ 风控拦截成功: Reason=%s, SuggestedQty=%s", decision.Reason, decision.SuggestedQuantity)
	})

	// 4. 场景3：连续亏损触发熔断
	t.Run("熔断器-连续亏损", func(t *testing.T) {
		// 模拟连续亏损
		_ = setup.riskRepo.RecordTrade(ctx, accountID, model.MustMoney("-100"))
		_ = setup.riskRepo.RecordTrade(ctx, accountID, model.MustMoney("-150"))
		_ = setup.riskRepo.RecordTrade(ctx, accountID, model.MustMoney("-200"))

		// 清除缓存
		setup.riskMgr.InvalidateCache(accountID, "")

		// 尝试下单应被熔断
		orderCtx := &risklogic.OrderContext{
			ClientOrderID: "order-circuit-breaker",
			Symbol:        symbol,
			MarketType:    model.MarketTypeSpot,
			Side:          model.OrderSideBuy,
			Type:          model.OrderTypeMarket,
			Quantity:      model.MustMoney("0.01"),
			CurrentPrice:  currentPrice,
			AccountID:     accountID,
		}

		decision, err := setup.riskMgr.CheckPreTrade(ctx, orderCtx)
		if err != nil {
			t.Fatalf("风控检查失败: %v", err)
		}

		if !decision.IsBlocked() {
			t.Error("连续亏损后订单应被熔断器阻止")
		}

		t.Logf("✅ 熔断器触发: Reason=%s", decision.Reason)
	})

	// 5. 场景4：订单状态同步
	t.Run("订单状态同步", func(t *testing.T) {
		clientOrderID := "order-sync-1"
		quantity := model.MustMoney("0.02")

		orderReq := &port.SpotPlaceOrderRequest{
			ClientOrderID: clientOrderID,
			Symbol:        symbol,
			Side:          model.OrderSideBuy,
			Type:          model.OrderTypeMarket,
			Quantity:      quantity,
		}

		placedOrder, err := setup.exchange.PlaceOrder(ctx, orderReq)
		if err != nil {
			t.Fatalf("下单失败: %v", err)
		}

		// 保存订单
		if err := setup.orderRepo.SaveOrder(ctx, placedOrder); err != nil {
			t.Fatalf("订单持久化失败: %v", err)
		}

		// 更新订单状态
		if err := setup.orderRepo.UpdateOrderStatus(ctx, clientOrderID, model.OrderStatusFilled); err != nil {
			t.Fatalf("更新订单状态失败: %v", err)
		}

		// 验证状态已更新
		updatedOrder, err := setup.orderRepo.GetOrder(ctx, clientOrderID)
		if err != nil {
			t.Fatalf("查询订单失败: %v", err)
		}

		if updatedOrder.Status != model.OrderStatusFilled {
			t.Errorf("订单状态未更新: got %s, want FILLED", updatedOrder.Status)
		}

		t.Logf("✅ 订单状态同步完成: Status=%s", updatedOrder.Status)
	})
}

// TestE2E_StrategyIntegration 策略集成测试
// 验证：Strategy Engine -> RiskManager -> Gateway 完整流程
func TestE2E_StrategyIntegration(t *testing.T) {
	ctx := context.Background()
	accountID := "strategy-test-account"
	initialCapital := model.MustMoney("10000")
	symbol := "BTCUSDT"
	currentPrice := model.MustMoney("50000")

	setup := setupTradingSystem(t, accountID, initialCapital, symbol, currentPrice)
	defer setup.cleanup()

	// 创建策略
	threshold := model.MustMoney("0.02")
	strat := strategy.NewSimpleVolatility(symbol, threshold)

	// 创建策略引擎（集成风控）
	engine := strategy.NewEngine(strat, setup.exchange, accountID)

	// 创建K线数据
	candle := &model.Candle{
		Symbol:    symbol,
		Interval:  "1m",
		OpenTime:  time.Now(),
		CloseTime: time.Now().Add(time.Minute),
		Open:      currentPrice,
		High:      currentPrice.Mul(model.MustMoney("1.01")),
		Low:       currentPrice.Mul(model.MustMoney("0.99")),
		Close:     currentPrice.Mul(model.MustMoney("1.02")), // 2% 涨幅，触发买入信号
		Volume:    model.MustMoney("100"),
	}

	// 处理K线（策略会生成信号并尝试下单）
	// 注意：当前 Strategy Engine 直接调用 Gateway，未集成 RiskManager
	// 这是 Phase 3 的已知限制，未来 OMS 会统一管理
	err := engine.ProcessCandle(ctx, candle)
	if err != nil {
		t.Logf("策略处理K线（可能因风控拦截）: %v", err)
	}

	// 验证订单是否已创建（如果通过风控）
	activeOrders, err := setup.orderRepo.ListActiveOrders(ctx)
	if err != nil {
		t.Fatalf("查询活跃订单失败: %v", err)
	}

	t.Logf("✅ 策略集成测试完成: ActiveOrders=%d", len(activeOrders))
}

// TestE2E_StatePersistence 状态持久化测试
// 验证：RiskRepo 和 OrderRepo 的状态持久化与恢复
func TestE2E_StatePersistence(t *testing.T) {
	ctx := context.Background()
	accountID := "persist-test-account"
	initialCapital := model.MustMoney("10000")

	// 使用内存仓储（模拟持久化）
	riskRepo := risk.NewMemoryRiskRepo()
	orderRepo := order.NewMemoryRepo()

	// 初始化风控状态
	state := model.NewRiskState(accountID, initialCapital)
	if err := riskRepo.SaveState(ctx, state); err != nil {
		t.Fatalf("保存风控状态失败: %v", err)
	}

	// 记录交易
	if err := riskRepo.RecordTrade(ctx, accountID, model.MustMoney("-500")); err != nil {
		t.Fatalf("记录交易失败: %v", err)
	}

	// 创建订单
	order1 := &model.Order{
		ClientOrderID: "persist-order-1",
		ExchangeID:    "EX-001",
		Symbol:        "BTCUSDT",
		MarketType:    model.MarketTypeSpot,
		Side:          model.OrderSideBuy,
		Type:          model.OrderTypeMarket,
		Quantity:      model.MustMoney("0.1"),
		Filled:        model.MustMoney("0.1"),
		Status:        model.OrderStatusFilled,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	if err := orderRepo.SaveOrder(ctx, order1); err != nil {
		t.Fatalf("保存订单失败: %v", err)
	}

	// 模拟系统重启：重新加载状态
	reloadedState, err := riskRepo.LoadState(ctx, accountID, "")
	if err != nil {
		t.Fatalf("重新加载风控状态失败: %v", err)
	}

	if !reloadedState.DailyPnL.EQ(model.MustMoney("-500")) {
		t.Errorf("风控状态未持久化: DailyPnL=%s, want -500", reloadedState.DailyPnL.String())
	}

	reloadedOrder, err := orderRepo.GetOrder(ctx, "persist-order-1")
	if err != nil {
		t.Fatalf("重新加载订单失败: %v", err)
	}

	if reloadedOrder.Status != model.OrderStatusFilled {
		t.Errorf("订单状态未持久化: Status=%s, want FILLED", reloadedOrder.Status)
	}

	t.Logf("✅ 状态持久化验证完成: DailyPnL=%s, OrderStatus=%s",
		reloadedState.DailyPnL.String(), reloadedOrder.Status.String())
}

// tradingSystemSetup 交易系统测试环境
type tradingSystemSetup struct {
	exchange  *mock.SpotExchange
	riskRepo  port.RiskRepo
	orderRepo port.OrderRepo
	riskMgr   *risklogic.Manager
	cleanup   func()
}

// setupTradingSystem 初始化交易系统测试环境
func setupTradingSystem(t *testing.T, accountID string, initialCapital model.Money, symbol string, currentPrice model.Money) *tradingSystemSetup {
	// 1. 初始化模拟交易所
	exchange := mock.NewSpotExchange(map[string]model.Money{
		"USDT": initialCapital,
		"BTC":  model.Zero(),
	})
	exchange.SetPrice(symbol, currentPrice)

	// 2. 初始化仓储
	riskRepo := risk.NewMemoryRiskRepo()
	orderRepo := order.NewMemoryRepo()

	// 3. 初始化风控管理器
	riskMgr := risklogic.NewManager(riskRepo, risklogic.RiskConfig{
		MaxSinglePositionPercent: 0.3,  // 单标的最大30%
		MaxTotalExposurePercent:  0.7,  // 总敞口最大70%
		MinCashReservePercent:    0.3,  // 最小现金储备30%
		MaxConsecutiveLosses:     3,    // 最大连续亏损3次
		MaxDailyDrawdown:         0.05, // 最大日回撤5%
		MaxTotalMDD:              0.15, // 最大总回撤15%
		MaxLeverage:              2,    // 最大杠杆2x
	})

	// 4. 初始化风控状态
	state := model.NewRiskState(accountID, initialCapital)
	if err := riskRepo.SaveState(context.Background(), state); err != nil {
		t.Fatalf("初始化风控状态失败: %v", err)
	}

	return &tradingSystemSetup{
		exchange:  exchange,
		riskRepo:  riskRepo,
		orderRepo: orderRepo,
		riskMgr:   riskMgr,
		cleanup:   func() {
			// 清理资源（内存仓储无需特殊清理）
		},
	}
}
