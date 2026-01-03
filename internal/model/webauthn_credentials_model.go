package model

import "github.com/zeromicro/go-zero/core/stores/sqlx"

var _ WebauthnCredentialsModel = (*customWebauthnCredentialsModel)(nil)

type (
	// WebauthnCredentialsModel is an interface to be customized, add more methods here,
	// and implement the added methods in customWebauthnCredentialsModel.
	WebauthnCredentialsModel interface {
		webauthnCredentialsModel
		withSession(session sqlx.Session) WebauthnCredentialsModel
	}

	customWebauthnCredentialsModel struct {
		*defaultWebauthnCredentialsModel
	}
)

// NewWebauthnCredentialsModel returns a model for the database table.
func NewWebauthnCredentialsModel(conn sqlx.SqlConn) WebauthnCredentialsModel {
	return &customWebauthnCredentialsModel{
		defaultWebauthnCredentialsModel: newWebauthnCredentialsModel(conn),
	}
}

func (m *customWebauthnCredentialsModel) withSession(session sqlx.Session) WebauthnCredentialsModel {
	return NewWebauthnCredentialsModel(sqlx.NewSqlConnFromSession(session))
}
