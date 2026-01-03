package passkey

import (
	"context"

	"github.com/iluyuns/alpha-trade/internal/svc"
	"github.com/iluyuns/alpha-trade/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type PasskeyAddBeginLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewPasskeyAddBeginLogic(ctx context.Context, svcCtx *svc.ServiceContext) PasskeyAddBeginLogic {
	return PasskeyAddBeginLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *PasskeyAddBeginLogic) PasskeyAddBegin(req *types.AddBeginRequest) (resp *types.AddBeginResponse, err error) {
	// todo: add your logic here and delete this line

	return
}
