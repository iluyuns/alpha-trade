package svc

import (
	"context"
	"fmt"
	"sync"

	"github.com/iluyuns/alpha-trade/internal/gateway/binance"
	"github.com/iluyuns/alpha-trade/internal/strategy"
	"github.com/iluyuns/alpha-trade/internal/domain/model"
	"github.com/zeromicro/go-zero/core/logx"
)

// TradingLoop 交易循环
// 负责订阅 WebSocket 行情，并将数据传递给策略引擎处理
type TradingLoop struct {
	wsClient      *binance.WSClient
	strategyEngine *strategy.Engine
	symbols       []string
	interval      string
	ctx           context.Context
	cancel        context.CancelFunc
	wg            sync.WaitGroup
	started       bool
	mu            sync.RWMutex
}

// NewTradingLoop 创建交易循环
func NewTradingLoop(
	wsClient *binance.WSClient,
	strategyEngine *strategy.Engine,
	symbols []string,
	interval string,
) *TradingLoop {
	ctx, cancel := context.WithCancel(context.Background())
	return &TradingLoop{
		wsClient:      wsClient,
		strategyEngine: strategyEngine,
		symbols:       symbols,
		interval:      interval,
		ctx:           ctx,
		cancel:        cancel,
		started:       false,
	}
}

// Start 启动交易循环
func (tl *TradingLoop) Start(ctx context.Context) error {
	tl.mu.Lock()
	defer tl.mu.Unlock()

	if tl.started {
		return nil // 已经启动
	}

	if len(tl.symbols) == 0 {
		return fmt.Errorf("no symbols configured for trading")
	}

	logx.Infof("Starting trading loop: symbols=%v, interval=%s", tl.symbols, tl.interval)

	// 订阅 K线数据
	candleCh, err := tl.wsClient.SubscribeKLines(ctx, tl.symbols, tl.interval)
	if err != nil {
		return fmt.Errorf("subscribe klines failed: %w", err)
	}

	tl.started = true

	// 启动处理协程
	tl.wg.Add(1)
	go tl.processCandles(candleCh)

	logx.Infof("Trading loop started successfully")
	return nil
}

// Stop 停止交易循环（优雅关闭）
func (tl *TradingLoop) Stop() {
	tl.mu.Lock()
	defer tl.mu.Unlock()

	if !tl.started {
		return
	}

	logx.Infof("Stopping trading loop...")

	// 取消上下文，停止接收新消息
	tl.cancel()

	// 等待处理协程完成
	tl.wg.Wait()

	tl.started = false
	logx.Infof("Trading loop stopped")
}

// processCandles 处理 K线数据
func (tl *TradingLoop) processCandles(ch <-chan *model.Candle) {
	defer tl.wg.Done()

	for {
		select {
		case <-tl.ctx.Done():
			logx.Infof("Trading loop context cancelled, stopping candle processing")
			return
		case candle, ok := <-ch:
			if !ok {
				logx.Errorf("Candle channel closed, trading loop will stop")
				return
			}

			// 处理 K线
			if err := tl.handleCandle(candle); err != nil {
				logx.Errorf("Error handling candle for %s: %v", candle.Symbol, err)
				// 继续处理下一个，不中断循环
			}
		}
	}
}

// handleCandle 处理单个 K线
func (tl *TradingLoop) handleCandle(candle *model.Candle) error {
	// 调用策略引擎处理 K线
	if err := tl.strategyEngine.ProcessCandle(tl.ctx, candle); err != nil {
		return fmt.Errorf("strategy engine process candle failed: %w", err)
	}

	return nil
}

// IsStarted 检查是否已启动
func (tl *TradingLoop) IsStarted() bool {
	tl.mu.RLock()
	defer tl.mu.RUnlock()
	return tl.started
}
