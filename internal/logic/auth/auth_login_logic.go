package auth

import (
	"context"

	"github.com/iluyuns/alpha-trade/internal/code"
	"github.com/iluyuns/alpha-trade/internal/model"
	"github.com/iluyuns/alpha-trade/internal/pkg/crypto"
	"github.com/iluyuns/alpha-trade/internal/pkg/ctxval"
	"github.com/iluyuns/alpha-trade/internal/pkg/jwt"
	"github.com/iluyuns/alpha-trade/internal/svc"
	"github.com/iluyuns/alpha-trade/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type AuthLoginLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAuthLoginLogic(ctx context.Context, svcCtx *svc.ServiceContext) AuthLoginLogic {
	return AuthLoginLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AuthLoginLogic) AuthLogin(req *types.LoginRequest) (resp *types.LoginResponse, err error) {
	user, err := l.svcCtx.UsersModel.FindOneByUsername(l.ctx, req.Username)
	if err != nil {
		l.recordAccessLog(0, "LOGIN", "FAIL", "USER_NOT_FOUND")
		return nil, err
	}
	// 校验密码 (Argon2id 生产级校验)
	match, err := crypto.VerifyPassword(req.Password, user.PasswordHash)
	if err != nil || !match {
		l.recordAccessLog(user.Id, "LOGIN", "FAIL", "INVALID_CREDENTIALS")
		return nil, code.New(code.ErrUsernameOrPasswordIncorrect)
	}

	// 认证 1 (Base Auth): 生成待 MFA 验证的临时 Token
	token, err := jwt.GenerateTokenWithIp(
		l.svcCtx.Config.Auth.AuthSecret,
		user.Id,
		jwt.ScopeBaseAuth,
		l.svcCtx.Config.Auth.BaseExpire,
		ctxval.GetIP(l.ctx),
	)
	if err != nil {
		l.recordAccessLog(user.Id, "LOGIN", "FAIL", "TOKEN_GEN_ERROR")
		return nil, err
	}

	l.recordAccessLog(user.Id, "LOGIN", "SUCCESS", "")

	return &types.LoginResponse{
		Status: "success",
		Token:  token,
		User: types.UserInfo{
			Id:          user.Id,
			Username:    user.Username,
			DisplayName: user.DisplayName,
			Avatar:      user.Avatar,
		},
	}, nil
}

func (l *AuthLoginLogic) recordAccessLog(uid int64, action, status, reason string) {
	_, _ = l.svcCtx.UserAccessLogsModel.Insert(l.ctx, &model.UserAccessLogs{
		UserId:    uid,
		IpAddress: ctxval.GetIP(l.ctx),
		UserAgent: ctxval.GetUA(l.ctx),
		Action:    action,
		Status:    status,
		Reason:    reason,
	})
}
