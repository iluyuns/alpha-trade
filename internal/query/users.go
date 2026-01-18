package query

import (
	"context"
	"time"
)

// UsersCustom is the custom extension for Users.
// Add your custom methods here. This file will NOT be overwritten on regeneration.
type UsersCustom struct {
	*usersDo
}

// NewUsers creates a new Users data accessor with custom methods.
// Use this constructor to get both generated and custom methods.
func NewUsers(db Executor) *UsersCustom {
	return &UsersCustom{
		usersDo: users.WithDB(db).(*usersDo),
	}
}

// FindByUsername 根据用户名查询用户
func (c *UsersCustom) FindByUsername(ctx context.Context, username string) (*Users, error) {
	users, err := c.Where(usersField.Username.Eq(username)).Find(ctx)
	if err != nil {
		return nil, err
	}
	if len(users) == 0 {
		return nil, ErrRecordNotFound
	}
	return users[0], nil
}

// FindByOAuth 根据第三方账号查询用户
func (c *UsersCustom) FindByOAuth(ctx context.Context, provider string, oauthID string) (*Users, error) {
	var cond WhereCondition
	switch provider {
	case "github":
		cond = usersField.GithubID.Eq(oauthID)
	case "google":
		cond = usersField.GoogleID.Eq(oauthID)
	default:
		return nil, ErrRecordNotFound
	}

	result, err := c.Where(cond).Find(ctx)
	if err != nil {
		return nil, err
	}
	if len(result) == 0 {
		return nil, ErrRecordNotFound
	}
	return result[0], nil
}

// GetRevokedAt 获取用户的撤销时间（用于 token 撤销检查）
// 实现 RevocationStore 接口
func (c *UsersCustom) GetRevokedAt(ctx context.Context, userID int64) (time.Time, error) {
	user, err := c.FindByPK(ctx, userID)
	if err != nil {
		return time.Time{}, err
	}
	return user.RevokedAt, nil
}

// UpdateRevokedAt 更新用户的撤销时间
// 实现 RevocationStore 接口
func (c *UsersCustom) UpdateRevokedAt(ctx context.Context, userID int64, revokedAt time.Time) error {
	user, err := c.FindByPK(ctx, userID)
	if err != nil {
		return err
	}
	user.RevokedAt = revokedAt
	return c.UpdateByPK(ctx, user)
}
