package model

import "github.com/zeromicro/go-zero/core/stores/sqlx"

var _ AssetSnapshotsModel = (*customAssetSnapshotsModel)(nil)

type (
	// AssetSnapshotsModel is an interface to be customized, add more methods here,
	// and implement the added methods in customAssetSnapshotsModel.
	AssetSnapshotsModel interface {
		assetSnapshotsModel
		withSession(session sqlx.Session) AssetSnapshotsModel
	}

	customAssetSnapshotsModel struct {
		*defaultAssetSnapshotsModel
	}
)

// NewAssetSnapshotsModel returns a model for the database table.
func NewAssetSnapshotsModel(conn sqlx.SqlConn) AssetSnapshotsModel {
	return &customAssetSnapshotsModel{
		defaultAssetSnapshotsModel: newAssetSnapshotsModel(conn),
	}
}

func (m *customAssetSnapshotsModel) withSession(session sqlx.Session) AssetSnapshotsModel {
	return NewAssetSnapshotsModel(sqlx.NewSqlConnFromSession(session))
}
