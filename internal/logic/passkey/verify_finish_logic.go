package passkey

import (
	"context"

	"github.com/iluyuns/alpha-trade/internal/svc"
	"github.com/iluyuns/alpha-trade/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type VerifyFinishLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewVerifyFinishLogic(ctx context.Context, svcCtx *svc.ServiceContext) VerifyFinishLogic {
	return VerifyFinishLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *VerifyFinishLogic) VerifyFinish(req *types.VerifyFinishRequest) (resp *types.VerifyFinishResponse, err error) {
	// todo: add your logic here and delete this line

	return
}
