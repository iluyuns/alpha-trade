package model

import "github.com/zeromicro/go-zero/core/stores/sqlx"

var _ StrategyConfigsModel = (*customStrategyConfigsModel)(nil)

type (
	// StrategyConfigsModel is an interface to be customized, add more methods here,
	// and implement the added methods in customStrategyConfigsModel.
	StrategyConfigsModel interface {
		strategyConfigsModel
		withSession(session sqlx.Session) StrategyConfigsModel
	}

	customStrategyConfigsModel struct {
		*defaultStrategyConfigsModel
	}
)

// NewStrategyConfigsModel returns a model for the database table.
func NewStrategyConfigsModel(conn sqlx.SqlConn) StrategyConfigsModel {
	return &customStrategyConfigsModel{
		defaultStrategyConfigsModel: newStrategyConfigsModel(conn),
	}
}

func (m *customStrategyConfigsModel) withSession(session sqlx.Session) StrategyConfigsModel {
	return NewStrategyConfigsModel(sqlx.NewSqlConnFromSession(session))
}
