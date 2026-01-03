package model

import "github.com/zeromicro/go-zero/core/stores/sqlx"

var _ ExchangeAccountsModel = (*customExchangeAccountsModel)(nil)

type (
	// ExchangeAccountsModel is an interface to be customized, add more methods here,
	// and implement the added methods in customExchangeAccountsModel.
	ExchangeAccountsModel interface {
		exchangeAccountsModel
		withSession(session sqlx.Session) ExchangeAccountsModel
	}

	customExchangeAccountsModel struct {
		*defaultExchangeAccountsModel
	}
)

// NewExchangeAccountsModel returns a model for the database table.
func NewExchangeAccountsModel(conn sqlx.SqlConn) ExchangeAccountsModel {
	return &customExchangeAccountsModel{
		defaultExchangeAccountsModel: newExchangeAccountsModel(conn),
	}
}

func (m *customExchangeAccountsModel) withSession(session sqlx.Session) ExchangeAccountsModel {
	return NewExchangeAccountsModel(sqlx.NewSqlConnFromSession(session))
}
