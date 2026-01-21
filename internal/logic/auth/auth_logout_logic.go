package auth

import (
	"context"
	"time"

	"github.com/iluyuns/alpha-trade/internal/pkg/ctxval"
	"github.com/iluyuns/alpha-trade/internal/svc"
	"github.com/iluyuns/alpha-trade/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type AuthLogoutLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAuthLogoutLogic(ctx context.Context, svcCtx *svc.ServiceContext) AuthLogoutLogic {
	return AuthLogoutLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AuthLogoutLogic) AuthLogout() (resp *types.LogoutResponse, err error) {
	uid, _ := l.ctx.Value("uid").(int64)

	// 更新内存中的撤销状态
	now := time.Now()
	l.svcCtx.RevocationManager.Revoke(l.ctx, uid, now)

	// 记录审计日志
	_ = l.svcCtx.AuditLogs.RecordAction(
		l.ctx,
		uid,
		ctxval.GetIP(l.ctx),
		"LOGOUT",
		"SUCCESS",
		"",
		"",
		false,
	)

	return &types.LogoutResponse{
		Success: true,
	}, nil
}
