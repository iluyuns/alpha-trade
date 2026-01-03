package model

import "github.com/zeromicro/go-zero/core/stores/sqlx"

var _ ExecutionsModel = (*customExecutionsModel)(nil)

type (
	// ExecutionsModel is an interface to be customized, add more methods here,
	// and implement the added methods in customExecutionsModel.
	ExecutionsModel interface {
		executionsModel
		withSession(session sqlx.Session) ExecutionsModel
	}

	customExecutionsModel struct {
		*defaultExecutionsModel
	}
)

// NewExecutionsModel returns a model for the database table.
func NewExecutionsModel(conn sqlx.SqlConn) ExecutionsModel {
	return &customExecutionsModel{
		defaultExecutionsModel: newExecutionsModel(conn),
	}
}

func (m *customExecutionsModel) withSession(session sqlx.Session) ExecutionsModel {
	return NewExecutionsModel(sqlx.NewSqlConnFromSession(session))
}
