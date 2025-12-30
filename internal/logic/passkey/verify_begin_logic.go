package passkey

import (
	"context"

	"github.com/iluyuns/alpha-trade/internal/svc"
	"github.com/iluyuns/alpha-trade/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type VerifyBeginLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewVerifyBeginLogic(ctx context.Context, svcCtx *svc.ServiceContext) VerifyBeginLogic {
	return VerifyBeginLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *VerifyBeginLogic) VerifyBegin(req *types.VerifyBeginRequest) (resp *types.VerifyBeginResponse, err error) {
	// todo: add your logic here and delete this line

	return
}
