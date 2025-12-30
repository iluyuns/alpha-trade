package passkey

import (
	"context"

	"github.com/iluyuns/alpha-trade/internal/svc"
	"github.com/iluyuns/alpha-trade/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type AddBeginLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAddBeginLogic(ctx context.Context, svcCtx *svc.ServiceContext) AddBeginLogic {
	return AddBeginLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AddBeginLogic) AddBegin(req *types.AddBeginRequest) (resp *types.AddBeginResponse, err error) {
	// todo: add your logic here and delete this line

	return
}
