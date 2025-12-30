package passkey

import (
	"context"

	"github.com/iluyuns/alpha-trade/internal/svc"
	"github.com/iluyuns/alpha-trade/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type AddFinishLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAddFinishLogic(ctx context.Context, svcCtx *svc.ServiceContext) AddFinishLogic {
	return AddFinishLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AddFinishLogic) AddFinish(req *types.AddFinishRequest) (resp *types.AddFinishResponse, err error) {
	// todo: add your logic here and delete this line

	return
}
