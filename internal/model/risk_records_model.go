package model

import "github.com/zeromicro/go-zero/core/stores/sqlx"

var _ RiskRecordsModel = (*customRiskRecordsModel)(nil)

type (
	// RiskRecordsModel is an interface to be customized, add more methods here,
	// and implement the added methods in customRiskRecordsModel.
	RiskRecordsModel interface {
		riskRecordsModel
		withSession(session sqlx.Session) RiskRecordsModel
	}

	customRiskRecordsModel struct {
		*defaultRiskRecordsModel
	}
)

// NewRiskRecordsModel returns a model for the database table.
func NewRiskRecordsModel(conn sqlx.SqlConn) RiskRecordsModel {
	return &customRiskRecordsModel{
		defaultRiskRecordsModel: newRiskRecordsModel(conn),
	}
}

func (m *customRiskRecordsModel) withSession(session sqlx.Session) RiskRecordsModel {
	return NewRiskRecordsModel(sqlx.NewSqlConnFromSession(session))
}
