package passkey

import (
	"context"
	"errors"

	"github.com/iluyuns/alpha-trade/internal/svc"
	"github.com/iluyuns/alpha-trade/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type PasskeyListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewPasskeyListLogic(ctx context.Context, svcCtx *svc.ServiceContext) PasskeyListLogic {
	return PasskeyListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *PasskeyListLogic) PasskeyList() (resp *types.ListResponse, err error) {
	// Passkey 功能暂未实现 (Phase 4)
	return nil, errors.New("passkey feature not implemented yet")
}
