package risk

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/iluyuns/alpha-trade/internal/domain/model"
)

// RedisRepo Redis 风控仓储（实现 port.RiskRepo 接口）
// 使用 JSON 序列化存储 RiskState，支持 TTL 过期
type RedisRepo struct {
	client *redis.Client
	ttl    time.Duration // 状态过期时间（默认 24 小时）
}

// NewRedisRepo 创建 Redis 风控仓储
func NewRedisRepo(client *redis.Client) *RedisRepo {
	return &RedisRepo{
		client: client,
		ttl:    24 * time.Hour, // 默认 24 小时过期
	}
}

// NewRedisRepoWithTTL 创建 Redis 风控仓储（自定义 TTL）
func NewRedisRepoWithTTL(client *redis.Client, ttl time.Duration) *RedisRepo {
	return &RedisRepo{
		client: client,
		ttl:    ttl,
	}
}

// makeKey 生成 Redis 键
func (r *RedisRepo) makeKey(accountID, symbol string) string {
	if symbol == "" {
		return fmt.Sprintf("risk:state:%s", accountID)
	}
	return fmt.Sprintf("risk:state:%s:%s", accountID, symbol)
}

// LoadState 加载风控状态
func (r *RedisRepo) LoadState(ctx context.Context, accountID, symbol string) (*model.RiskState, error) {
	key := r.makeKey(accountID, symbol)

	// 从 Redis 读取 JSON
	data, err := r.client.Get(ctx, key).Result()
	if err == redis.Nil {
		// 记录不存在，返回新状态
		return model.NewRiskState(accountID, model.Zero()), nil
	}
	if err != nil {
		return nil, fmt.Errorf("load risk state from redis failed: %w", err)
	}

	// 反序列化
	var state model.RiskState
	if err := json.Unmarshal([]byte(data), &state); err != nil {
		return nil, fmt.Errorf("unmarshal risk state failed: %w", err)
	}

	return &state, nil
}

// SaveState 保存风控状态（幂等）
func (r *RedisRepo) SaveState(ctx context.Context, state *model.RiskState) error {
	key := r.makeKey(state.AccountID, state.Symbol)

	// 序列化为 JSON
	data, err := json.Marshal(state)
	if err != nil {
		return fmt.Errorf("marshal risk state failed: %w", err)
	}

	// 写入 Redis（带 TTL）
	if err := r.client.Set(ctx, key, data, r.ttl).Err(); err != nil {
		return fmt.Errorf("save risk state to redis failed: %w", err)
	}

	return nil
}

// UpdateEquity 原子更新净值
func (r *RedisRepo) UpdateEquity(ctx context.Context, accountID string, newEquity model.Money) error {
	// 先加载状态
	state, err := r.LoadState(ctx, accountID, "")
	if err != nil {
		return err
	}

	// 更新净值
	state.UpdateEquity(newEquity)

	// 保存
	return r.SaveState(ctx, state)
}

// RecordTrade 记录交易（更新当日统计）
func (r *RedisRepo) RecordTrade(ctx context.Context, accountID string, pnl model.Money) error {
	// 先加载状态
	state, err := r.LoadState(ctx, accountID, "")
	if err != nil {
		return err
	}

	// 更新状态
	state.DailyPnL = state.DailyPnL.Add(pnl)
	state.DailyTradeCount++

	// 更新连续亏损
	if pnl.IsNegative() {
		state.RecordLoss()
	} else if pnl.IsPositive() {
		state.ResetConsecutiveLosses()
	}

	// 保存
	return r.SaveState(ctx, state)
}

// OpenCircuitBreaker 打开熔断器
func (r *RedisRepo) OpenCircuitBreaker(ctx context.Context, accountID string, duration int64) error {
	// 先加载状态
	state, err := r.LoadState(ctx, accountID, "")
	if err != nil {
		return err
	}

	// 打开熔断器
	state.OpenCircuitBreaker(time.Duration(duration) * time.Second)

	// 保存
	return r.SaveState(ctx, state)
}

// CloseCircuitBreaker 关闭熔断器
func (r *RedisRepo) CloseCircuitBreaker(ctx context.Context, accountID string) error {
	// 先加载状态
	state, err := r.LoadState(ctx, accountID, "")
	if err != nil {
		return err
	}

	// 关闭熔断器
	state.CloseCircuitBreaker()

	// 保存
	return r.SaveState(ctx, state)
}

// IsCircuitBreakerOpen 检查熔断器状态
func (r *RedisRepo) IsCircuitBreakerOpen(ctx context.Context, accountID string) (bool, error) {
	state, err := r.LoadState(ctx, accountID, "")
	if err != nil {
		return false, err
	}

	if !state.CircuitBreakerOpen {
		return false, nil
	}

	// 检查是否过期
	if state.CircuitBreakerUntil > 0 {
		if time.Now().Unix() >= state.CircuitBreakerUntil {
			// 自动关闭过期的熔断器
			_ = r.CloseCircuitBreaker(ctx, accountID)
			return false, nil
		}
	}

	return true, nil
}
