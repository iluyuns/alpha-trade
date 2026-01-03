package svc

import (
	"github.com/iluyuns/alpha-trade/internal/config"
	"github.com/iluyuns/alpha-trade/internal/middleware"
	"github.com/iluyuns/alpha-trade/internal/model"
	"github.com/iluyuns/alpha-trade/internal/pkg/email"
	"github.com/iluyuns/alpha-trade/internal/pkg/revocation"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"github.com/zeromicro/go-zero/rest"
)

type ServiceContext struct {
	Config                   config.Config
	Auth                     rest.Middleware
	MFA                      rest.Middleware
	MFAStepUp                rest.Middleware
	Email                    email.EmailService
	Conn                     sqlx.SqlConn
	RevocationManager        revocation.RevocationManager
	UsersModel               model.UsersModel
	WebauthnCredentialsModel model.WebauthnCredentialsModel
	UserAccessLogsModel      model.UserAccessLogsModel
}

func NewServiceContext(c config.Config) (*ServiceContext, error) {
	conn := sqlx.NewSqlConn("postgres", c.Database.DataSource)
	userAccessModel := model.NewUserAccessLogsModel(conn)
	usersModel := model.NewUsersModel(conn)
	revocationManager, err := revocation.NewCachedRevocationManager(usersModel)
	if err != nil {
		return nil, err
	}

	return &ServiceContext{
		Config:                   c,
		Auth:                     middleware.NewAuthMiddleware(c.Auth.AuthSecret, userAccessModel, revocationManager).Handle, // 基础/MFA 认证密钥
		MFA:                      middleware.NewMFAMiddleware().Handle,                                                       // MFA 状态校验
		MFAStepUp:                middleware.NewMFAStepUpMiddleware(c.Auth.SudoSecret).Handle,                                // 提级认证校验
		Email:                    email.NewAWSSES(&c.AWS),
		Conn:                     conn,
		RevocationManager:        revocationManager,
		UsersModel:               usersModel,
		WebauthnCredentialsModel: model.NewWebauthnCredentialsModel(conn),
		UserAccessLogsModel:      userAccessModel,
	}, nil
}
