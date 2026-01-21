package passkey

import (
	"context"
	"errors"

	"github.com/iluyuns/alpha-trade/internal/svc"
	"github.com/iluyuns/alpha-trade/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type PasskeyAddFinishLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewPasskeyAddFinishLogic(ctx context.Context, svcCtx *svc.ServiceContext) PasskeyAddFinishLogic {
	return PasskeyAddFinishLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *PasskeyAddFinishLogic) PasskeyAddFinish(req *types.AddFinishRequest) (resp *types.AddFinishResponse, err error) {
	// Passkey 功能暂未实现 (Phase 4)
	return nil, errors.New("passkey feature not implemented yet")
}
