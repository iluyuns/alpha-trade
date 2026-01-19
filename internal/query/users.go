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

func (c *UsersCustom) FindByUsername(ctx context.Context, username string) (*Users, error) {
	return c.Where(c.Field.Username.Eq(username)).First(ctx)
}

// FindByOAuth finds a user by OAuth provider and provider ID
func (c *UsersCustom) FindByOAuth(ctx context.Context, provider, providerID string) (*Users, error) {
	switch provider {
	case "github":
		return c.Where(c.Field.GithubID.Eq(providerID)).First(ctx)
	case "google":
		return c.Where(c.Field.GoogleID.Eq(providerID)).First(ctx)
	default:
		return nil, ErrRecordNotFound
	}
}

// UpdateRevokedAt updates the revoked_at timestamp for a user
func (c *UsersCustom) UpdateRevokedAt(ctx context.Context, uid int64, revokedAt time.Time) error {
	_, err := c.Where(c.Field.ID.Eq(uid)).Update(ctx, map[string]interface{}{
		"revoked_at": revokedAt,
	})
	return err
}

// GetRevokedAt retrieves the revoked_at timestamp for a user
func (c *UsersCustom) GetRevokedAt(ctx context.Context, uid int64) (time.Time, error) {
	user, err := c.Where(c.Field.ID.Eq(uid)).Select(c.Field.RevokedAt).First(ctx)
	if err != nil {
		return time.Time{}, err
	}
	return user.RevokedAt, nil
}
