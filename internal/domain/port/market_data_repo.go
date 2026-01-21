package port

import (
	"context"

	"github.com/iluyuns/alpha-trade/internal/domain/model"
)

// MarketDataRepo 行情数据接口
// 支持实时订阅（WebSocket）与历史拉取（REST API）
// 回测时使用迭代式喂给器（HistoricalIterator）
type MarketDataRepo interface {
	// SubscribeTicks 订阅 Tick 数据流（实盘用）
	// 返回的 channel 持续推送行情，直到 context 取消
	SubscribeTicks(ctx context.Context, symbols []string) (<-chan *model.Tick, error)

	// SubscribeKLines 订阅 K线数据流
	SubscribeKLines(ctx context.Context, symbols []string, interval string) (<-chan *model.Candle, error)

	// GetHistoricalKLines 拉取历史 K线（回测用）
	// startTime 和 endTime 为 Unix 毫秒时间戳
	GetHistoricalKLines(ctx context.Context, symbol string, interval string, startTime, endTime int64) ([]*model.Candle, error)

	// GetLatestPrice 获取最新价格（快照）
	GetLatestPrice(ctx context.Context, symbol string) (model.Money, error)
}

// HistoricalIterator 历史数据迭代器（回测专用）
// 按事件时间顺序返回行情数据，确保回测时间正确性
type HistoricalIterator interface {
	// Next 获取下一个行情事件
	// 返回 nil 表示数据结束
	Next() (*model.Candle, error)

	// HasNext 是否还有数据
	HasNext() bool

	// CurrentTime 当前回测时间（事件时间）
	CurrentTime() int64
}
