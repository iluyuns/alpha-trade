package strategy

import (
	"context"
	"testing"
	"time"

	"github.com/iluyuns/alpha-trade/internal/domain/model"
)

func TestSimpleVolatility_OnCandle(t *testing.T) {
	ctx := context.Background()
	strategy := NewSimpleVolatility("BTCUSDT", model.MustMoney("0.02")) // 2% threshold

	tests := []struct {
		name          string
		candles       []*model.Candle
		wantSignal    Signal
		wantNilSignal bool
	}{
		{
			name: "初始化不产生信号",
			candles: []*model.Candle{
				{Symbol: "BTCUSDT", Close: model.MustMoney("50000"), OpenTime: time.Now()},
			},
			wantNilSignal: true,
		},
		{
			name: "波动小于阈值不产生信号",
			candles: []*model.Candle{
				{Symbol: "BTCUSDT", Close: model.MustMoney("50000"), OpenTime: time.Now()},
				{Symbol: "BTCUSDT", Close: model.MustMoney("50500"), OpenTime: time.Now()}, // 1% < 2%
			},
			wantNilSignal: true,
		},
		{
			name: "上涨超过阈值产生买入信号",
			candles: []*model.Candle{
				{Symbol: "BTCUSDT", Close: model.MustMoney("50000"), OpenTime: time.Now()},
				{Symbol: "BTCUSDT", Close: model.MustMoney("51100"), OpenTime: time.Now()}, // 2.2% > 2%
			},
			wantSignal: SignalBuy,
		},
		{
			name: "下跌超过阈值且有持仓产生卖出信号",
			candles: []*model.Candle{
				{Symbol: "BTCUSDT", Close: model.MustMoney("50000"), OpenTime: time.Now()},
				{Symbol: "BTCUSDT", Close: model.MustMoney("51100"), OpenTime: time.Now()}, // Buy
				{Symbol: "BTCUSDT", Close: model.MustMoney("50000"), OpenTime: time.Now()}, // 下跌 2.15% > 2%
			},
			wantSignal: SignalSell,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 重置策略状态
			strategy.lastPrice = model.Zero()
			strategy.positionQuantity = model.Zero()

			var lastSignal *TradeSignal
			for _, candle := range tt.candles {
				signal, err := strategy.OnCandle(ctx, candle)
				if err != nil {
					t.Fatalf("OnCandle failed: %v", err)
				}
				if signal != nil {
					lastSignal = signal
				}
			}

			if tt.wantNilSignal {
				if lastSignal != nil {
					t.Errorf("Expected nil signal, got %s", lastSignal.Signal)
				}
				return
			}

			if lastSignal == nil {
				t.Fatal("Expected signal, got nil")
			}

			if lastSignal.Signal != tt.wantSignal {
				t.Errorf("Signal = %s, want %s", lastSignal.Signal, tt.wantSignal)
			}

			t.Logf("Signal: %s, Reason: %s", lastSignal.Signal, lastSignal.Reason)
		})
	}
}
