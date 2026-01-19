package query

import (
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

// Executor is the sqlx connection for database operations
// Use go-zero's sqlx which provides QueryRowCtx, QueryRowsCtx, ExecCtx methods
type Executor = sqlx.SqlConn
