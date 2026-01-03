package model

import "github.com/zeromicro/go-zero/core/stores/sqlx"

var _ UserAccessLogsModel = (*customUserAccessLogsModel)(nil)

type (
	// UserAccessLogsModel is an interface to be customized, add more methods here,
	// and implement the added methods in customUserAccessLogsModel.
	UserAccessLogsModel interface {
		userAccessLogsModel
		withSession(session sqlx.Session) UserAccessLogsModel
	}

	customUserAccessLogsModel struct {
		*defaultUserAccessLogsModel
	}
)

// NewUserAccessLogsModel returns a model for the database table.
func NewUserAccessLogsModel(conn sqlx.SqlConn) UserAccessLogsModel {
	return &customUserAccessLogsModel{
		defaultUserAccessLogsModel: newUserAccessLogsModel(conn),
	}
}

func (m *customUserAccessLogsModel) withSession(session sqlx.Session) UserAccessLogsModel {
	return NewUserAccessLogsModel(sqlx.NewSqlConnFromSession(session))
}

