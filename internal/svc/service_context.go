package svc

import (
	"database/sql"

	"github.com/iluyuns/alpha-trade/internal/config"
	"github.com/iluyuns/alpha-trade/internal/middleware"
	"github.com/iluyuns/alpha-trade/internal/pkg/email"
	"github.com/iluyuns/alpha-trade/internal/pkg/revocation"
	"github.com/iluyuns/alpha-trade/internal/query"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"github.com/zeromicro/go-zero/rest"
)

type ServiceContext struct {
	Config            config.Config
	Auth              rest.Middleware
	MFA               rest.Middleware
	MFAStepUp         rest.Middleware
	Email             email.EmailService
	DB                *sql.DB
	RevocationManager revocation.RevocationManager

	// Query 访问器
	Users               *query.UsersCustom
	WebauthnCredentials *query.WebauthnCredentialsCustom
	AuditLogs           *query.AuditLogsCustom
	UserAccessLogs      *query.UserAccessLogsCustom
}

func (sc *ServiceContext) Close() error {
	// 关闭数据库连接
	if sc.DB != nil {
		return sc.DB.Close()
	}
	return nil
}

func NewServiceContext(c config.Config) (*ServiceContext, error) {
	// 打开数据库连接
	db, err := sql.Open("postgres", c.Database.DataSource)
	if err != nil {
		return nil, err
	}

	// 创建 sqlx 连接
	sqlConn := sqlx.NewSqlConnFromDB(db)

	// 初始化 Query 访问器
	usersQuery := query.NewUsers(sqlConn)
	auditLogsQuery := query.NewAuditLogs(sqlConn)
	webauthnQuery := query.NewWebauthnCredentials(sqlConn)
	userAccessLogsQuery := query.NewUserAccessLogs(sqlConn)

	// 初始化撤销管理器
	revocationManager, err := revocation.NewCachedRevocationManager(usersQuery)
	if err != nil {
		return nil, err
	}

	return &ServiceContext{
		Config:            c,
		Auth:              middleware.NewAuthMiddleware(c.Auth.AuthSecret, userAccessLogsQuery, revocationManager).Handle, // 基础/MFA 认证密钥
		MFA:               middleware.NewMFAMiddleware().Handle,                                                           // MFA 状态校验
		MFAStepUp:         middleware.NewMFAStepUpMiddleware(c.Auth.SudoSecret).Handle,                                    // 提级认证校验
		Email:             email.NewAWSSES(&c.AWS),
		DB:                db,
		RevocationManager: revocationManager,

		// Query 访问器
		Users:               usersQuery,
		WebauthnCredentials: webauthnQuery,
		AuditLogs:           auditLogsQuery,
		UserAccessLogs:      userAccessLogsQuery,
	}, nil
}
