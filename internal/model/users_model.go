package model

import (
	"context"
	"fmt"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ UsersModel = (*customUsersModel)(nil)

type (
	// UsersModel is an interface to be customized, add more methods here,
	// and implement the added methods in customUsersModel.
	UsersModel interface {
		usersModel
		withSession(session sqlx.Session) UsersModel
		FindOneByUsername(ctx context.Context, username string) (*Users, error)
		FindOneByOAuth(ctx context.Context, provider, oauthId string) (*Users, error)
		UpdateRevokedAt(ctx context.Context, uid int64, revokedAt time.Time) error
		GetRevokedAt(ctx context.Context, uid int64) (time.Time, error)
	}

	customUsersModel struct {
		*defaultUsersModel
	}
)

// NewUsersModel returns a model for the database table.
func NewUsersModel(conn sqlx.SqlConn) UsersModel {
	return &customUsersModel{
		defaultUsersModel: newUsersModel(conn),
	}
}

func (m *customUsersModel) withSession(session sqlx.Session) UsersModel {
	return NewUsersModel(sqlx.NewSqlConnFromSession(session))
}

func (m *customUsersModel) FindOneByUsername(ctx context.Context, username string) (*Users, error) {
	var resp Users
	builder := m.selectBuilder().Where(squirrel.Eq{"username": username}).Limit(1)
	err := m.findWithAny(ctx, builder, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

func (m *customUsersModel) FindOneByOAuth(ctx context.Context, provider, oauthId string) (*Users, error) {
	var resp Users
	column := ""
	switch provider {
	case "github":
		column = "github_id"
	case "google":
		column = "google_id"
	default:
		return nil, fmt.Errorf("unsupported oauth provider: %s", provider)
	}

	builder := m.selectBuilder().Where(squirrel.Eq{column: oauthId}).Limit(1)
	err := m.findWithAny(ctx, builder, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

func (m *customUsersModel) UpdateRevokedAt(ctx context.Context, uid int64, revokedAt time.Time) error {
	query := fmt.Sprintf("update %s set revoked_at = $1 where id = $2", m.table)
	_, err := m.conn.ExecCtx(ctx, query, revokedAt, uid)
	return err
}

func (m *customUsersModel) GetRevokedAt(ctx context.Context, uid int64) (time.Time, error) {
	var revokedAt time.Time
	query := fmt.Sprintf("select revoked_at from %s where id = $1", m.table)
	err := m.conn.QueryRowCtx(ctx, &revokedAt, query, uid)
	if err != nil {
		return time.Time{}, err
	}
	return revokedAt, nil
}
