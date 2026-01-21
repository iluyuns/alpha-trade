package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/iluyuns/alpha-trade/internal/backtest/loader"
	"github.com/iluyuns/alpha-trade/internal/domain/model"
	"github.com/iluyuns/alpha-trade/internal/gateway/mock"
	"github.com/iluyuns/alpha-trade/internal/infra/risk"
	risklogic "github.com/iluyuns/alpha-trade/internal/logic/risk"
	"github.com/iluyuns/alpha-trade/internal/strategy"
)

var (
	csvFile   = flag.String("csv", "", "CSV数据文件路径")
	symbol    = flag.String("symbol", "BTCUSDT", "交易对")
	threshold = flag.String("threshold", "0.02", "波动阈值")
	capital   = flag.String("capital", "10000", "初始资金（USDT）")
)

func main() {
	flag.Parse()

	if *csvFile == "" {
		log.Fatal("请指定CSV文件路径: -csv /path/to/data.csv")
	}

	ctx := context.Background()

	// 1. 加载历史数据
	log.Printf("Loading data from %s...", *csvFile)
	dataLoader, err := loader.NewCsvLoader(*csvFile)
	if err != nil {
		log.Fatalf("Failed to load CSV: %v", err)
	}
	dataLoader.SetSymbol(*symbol)
	dataLoader.SetInterval("1m")
	log.Printf("Loaded %d candles", dataLoader.Count())

	// 2. 初始化模拟交易所
	initialCapital := model.MustMoney(*capital)
	exchange := mock.NewSpotExchange(map[string]model.Money{
		"USDT": initialCapital,
		"BTC":  model.Zero(),
	})

	// 3. 初始化风控系统
	riskRepo := risk.NewMemoryRiskRepo()
	_ = risklogic.NewManager(riskRepo, risklogic.RiskConfig{
		MaxSinglePositionPercent: 0.3,
		MaxTotalExposurePercent:  0.7,
		MinCashReservePercent:    0.3,
		MaxConsecutiveLosses:     3,
		MaxDailyDrawdown:         0.05,
		MaxTotalMDD:              0.15,
		MaxLeverage:              2,
	})

	// 初始化账户状态
	accountID := "backtest-account"
	accountState := model.NewRiskState(accountID, initialCapital)
	_ = riskRepo.SaveState(ctx, accountState)

	// 4. 初始化策略
	thresholdValue := model.MustMoney(*threshold)
	strat := strategy.NewSimpleVolatility(*symbol, thresholdValue)
	engine := strategy.NewEngine(strat, exchange, accountID)

	// 5. 回测循环
	log.Printf("Starting backtest with %s strategy (threshold: %s)", strat.Name(), *threshold)
	log.Println("=" + repeat("=", 60))

	stats := &BacktestStats{
		StartTime:    time.Now(),
		TotalCandles: dataLoader.Count(),
	}

	for dataLoader.HasNext() {
		candle, err := dataLoader.Next()
		if err != nil {
			log.Printf("Error reading candle: %v", err)
			continue
		}

		// 更新交易所价格
		exchange.SetPrice(candle.Symbol, candle.Close)

		// 风控检查（示例：检查每个信号）
		// 实际应该在策略引擎内部集成风控

		// 处理K线
		if err := engine.ProcessCandle(ctx, candle); err != nil {
			log.Printf("[%s] Error processing candle: %v",
				candle.OpenTime.Format("2006-01-02 15:04"), err)
			stats.Errors++
			continue
		}

		stats.ProcessedCandles++
	}

	// 6. 输出回测报告
	log.Println("=" + repeat("=", 60))
	log.Println("Backtest Complete!")
	log.Println("=" + repeat("=", 60))

	printReport(ctx, stats, exchange, accountID, initialCapital)
}

// BacktestStats 回测统计
type BacktestStats struct {
	StartTime        time.Time
	TotalCandles     int
	ProcessedCandles int
	Errors           int
}

// printReport 打印回测报告
func printReport(ctx context.Context, stats *BacktestStats, exchange *mock.SpotExchange, accountID string, initialCapital model.Money) {
	duration := time.Since(stats.StartTime)

	fmt.Println("\n📊 Backtest Report")
	fmt.Println(repeat("-", 60))

	// 时间统计
	fmt.Printf("Duration:         %s\n", duration.Round(time.Millisecond))
	fmt.Printf("Candles:          %d / %d processed\n", stats.ProcessedCandles, stats.TotalCandles)
	fmt.Printf("Errors:           %d\n", stats.Errors)
	fmt.Println()

	// 账户余额
	balances, _ := exchange.GetAllBalances(ctx)
	fmt.Println("💰 Final Balances:")
	totalValue := model.Zero()
	for _, bal := range balances {
		if bal.Total.IsPositive() {
			fmt.Printf("  %s: %s (Free: %s, Locked: %s)\n",
				bal.Asset, bal.Total.String(), bal.Free.String(), bal.Locked.String())
			totalValue = totalValue.Add(bal.Total)
		}
	}
	fmt.Println()

	// PnL计算（简化：仅USDT余额变化）
	usdtBal, _ := exchange.GetBalance(ctx, "USDT")
	pnl := usdtBal.Total.Sub(initialCapital)
	pnlPercent := pnl.Div(initialCapital)

	fmt.Println("📈 Performance:")
	fmt.Printf("  Initial Capital: %s USDT\n", initialCapital.String())
	fmt.Printf("  Final USDT:      %s\n", usdtBal.Total.String())
	fmt.Printf("  PnL:             %s (%.2f%%)\n", pnl.String(), pnlPercent.Float64()*100)
	fmt.Println()

	fmt.Println(repeat("=", 60))
}

// repeat 重复字符串
func repeat(s string, count int) string {
	result := ""
	for i := 0; i < count; i++ {
		result += s
	}
	return result
}
