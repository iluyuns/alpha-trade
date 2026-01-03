package passkey

import (
	"context"

	"github.com/iluyuns/alpha-trade/internal/svc"
	"github.com/iluyuns/alpha-trade/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type PasskeyVerifyFinishLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewPasskeyVerifyFinishLogic(ctx context.Context, svcCtx *svc.ServiceContext) PasskeyVerifyFinishLogic {
	return PasskeyVerifyFinishLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *PasskeyVerifyFinishLogic) PasskeyVerifyFinish(req *types.VerifyFinishRequest) (resp *types.VerifyFinishResponse, err error) {
	// todo: add your logic here and delete this line

	return
}
