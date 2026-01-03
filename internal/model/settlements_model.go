package model

import "github.com/zeromicro/go-zero/core/stores/sqlx"

var _ SettlementsModel = (*customSettlementsModel)(nil)

type (
	// SettlementsModel is an interface to be customized, add more methods here,
	// and implement the added methods in customSettlementsModel.
	SettlementsModel interface {
		settlementsModel
		withSession(session sqlx.Session) SettlementsModel
	}

	customSettlementsModel struct {
		*defaultSettlementsModel
	}
)

// NewSettlementsModel returns a model for the database table.
func NewSettlementsModel(conn sqlx.SqlConn) SettlementsModel {
	return &customSettlementsModel{
		defaultSettlementsModel: newSettlementsModel(conn),
	}
}

func (m *customSettlementsModel) withSession(session sqlx.Session) SettlementsModel {
	return NewSettlementsModel(sqlx.NewSqlConnFromSession(session))
}
