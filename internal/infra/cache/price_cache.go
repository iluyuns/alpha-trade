package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/iluyuns/alpha-trade/internal/domain/model"
	"github.com/iluyuns/alpha-trade/internal/domain/port"
)

// PriceCache Redis 价格缓存（装饰 MarketDataRepo，透明缓存 GetLatestPrice）
type PriceCache struct {
	upstream port.MarketDataRepo
	client   *redis.Client
	ttl      time.Duration
}

// NewPriceCache 创建价格缓存
// ttl 控制价格缓存时效；对于高频场景建议 1-5s
func NewPriceCache(upstream port.MarketDataRepo, client *redis.Client, ttl time.Duration) *PriceCache {
	return &PriceCache{
		upstream: upstream,
		client:   client,
		ttl:      ttl,
	}
}

func (c *PriceCache) priceKey(symbol string) string {
	return fmt.Sprintf("cache:price:%s", symbol)
}

// GetLatestPrice 优先从 Redis 读取，miss 时回源并写缓存
func (c *PriceCache) GetLatestPrice(ctx context.Context, symbol string) (model.Money, error) {
	key := c.priceKey(symbol)

	// 1. 尝试从缓存读取
	val, err := c.client.Get(ctx, key).Result()
	if err == nil && val != "" {
		price, err := model.NewMoney(val)
		if err == nil {
			return price, nil
		}
	}

	// 2. 缓存 miss，回源
	price, err := c.upstream.GetLatestPrice(ctx, symbol)
	if err != nil {
		return model.Zero(), err
	}

	// 3. 写缓存（best-effort，不阻塞主流程）
	_ = c.client.Set(ctx, key, price.String(), c.ttl).Err()

	return price, nil
}

// SetPrice 手动更新价格缓存（由 WebSocket Tick 回调驱动）
func (c *PriceCache) SetPrice(ctx context.Context, symbol string, price model.Money) error {
	return c.client.Set(ctx, c.priceKey(symbol), price.String(), c.ttl).Err()
}

// SubscribeTicks 透传到 upstream
func (c *PriceCache) SubscribeTicks(ctx context.Context, symbols []string) (<-chan *model.Tick, error) {
	tickCh, err := c.upstream.SubscribeTicks(ctx, symbols)
	if err != nil {
		return nil, err
	}

	// 包装 channel，自动更新价格缓存
	outCh := make(chan *model.Tick, 100)
	go func() {
		defer close(outCh)
		for tick := range tickCh {
			_ = c.SetPrice(ctx, tick.Symbol, tick.Price)
			select {
			case outCh <- tick:
			case <-ctx.Done():
				return
			}
		}
	}()

	return outCh, nil
}

// SubscribeKLines 透传到 upstream
func (c *PriceCache) SubscribeKLines(ctx context.Context, symbols []string, interval string) (<-chan *model.Candle, error) {
	candleCh, err := c.upstream.SubscribeKLines(ctx, symbols, interval)
	if err != nil {
		return nil, err
	}

	// 包装 channel，每根 K 线闭合时更新收盘价缓存
	outCh := make(chan *model.Candle, 100)
	go func() {
		defer close(outCh)
		for candle := range candleCh {
			_ = c.SetPrice(ctx, candle.Symbol, candle.Close)
			select {
			case outCh <- candle:
			case <-ctx.Done():
				return
			}
		}
	}()

	return outCh, nil
}

// GetHistoricalKLines 透传到 upstream（历史数据不缓存）
func (c *PriceCache) GetHistoricalKLines(ctx context.Context, symbol string, interval string, startTime, endTime int64) ([]*model.Candle, error) {
	return c.upstream.GetHistoricalKLines(ctx, symbol, interval, startTime, endTime)
}

// 确保 PriceCache 实现了 MarketDataRepo 接口
var _ port.MarketDataRepo = (*PriceCache)(nil)
