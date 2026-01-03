package revocation

import (
	"context"
	"strconv"
	"time"

	"github.com/zeromicro/go-zero/core/collection"
	"github.com/zeromicro/go-zero/core/syncx"
)

// RevocationStore 定义持久化层需要实现的接口
type RevocationStore interface {
	UpdateRevokedAt(ctx context.Context, uid int64, revokedAt time.Time) error
	GetRevokedAt(ctx context.Context, uid int64) (time.Time, error)
}

// RevocationManager 定义 UID 撤销管理器接口
type RevocationManager interface {
	Revoke(ctx context.Context, uid int64, revokedAt time.Time) error
	IsRevoked(ctx context.Context, uid int64, issuedAt time.Time) bool
}

// CachedRevocationManager 基于 go-zero cache + singleflight + store 的实现
type CachedRevocationManager struct {
	cache *collection.Cache
	sfg   syncx.SingleFlight
	store RevocationStore
}

func NewCachedRevocationManager(store RevocationStore) (*CachedRevocationManager, error) {
	// 使用 go-zero 的 Cache，设置 1 小时过期
	c, err := collection.NewCache(time.Hour)
	if err != nil {
		return nil, err
	}
	return &CachedRevocationManager{
		cache: c,
		sfg:   syncx.NewSingleFlight(),
		store: store,
	}, nil
}

func (m *CachedRevocationManager) Revoke(ctx context.Context, uid int64, revokedAt time.Time) error {
	// 1. 先写数据库 (Source of Truth)
	if err := m.store.UpdateRevokedAt(ctx, uid, revokedAt); err != nil {
		return err
	}

	// 2. 同步更新缓存
	m.cache.Set(strconv.FormatInt(uid, 10), revokedAt)
	return nil
}

func (m *CachedRevocationManager) IsRevoked(ctx context.Context, uid int64, issuedAt time.Time) bool {
	uidStr := strconv.FormatInt(uid, 10)

	// 1. 先查 Cache
	val, ok := m.cache.Get(uidStr)
	if ok {
		return issuedAt.Before(val.(time.Time))
	}

	// 2. Cache 未命中，使用 SingleFlight 处理并发穿透并回源 DB
	revokedAtVal, _ := m.sfg.Do(uidStr, func() (any, error) {
		// 从数据库加载
		revokedAt, err := m.store.GetRevokedAt(ctx, uid)
		if err != nil {
			return time.Time{}, err
		}

		m.cache.Set(uidStr, revokedAt)
		return revokedAt, nil
	})

	revokedAt, ok := revokedAtVal.(time.Time)
	if !ok || revokedAt.IsZero() {
		return false
	}

	return issuedAt.Before(revokedAt)
}
