package passkey

import (
	"context"

	"github.com/iluyuns/alpha-trade/internal/svc"
	"github.com/iluyuns/alpha-trade/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type PasskeyVerifyBeginLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewPasskeyVerifyBeginLogic(ctx context.Context, svcCtx *svc.ServiceContext) PasskeyVerifyBeginLogic {
	return PasskeyVerifyBeginLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *PasskeyVerifyBeginLogic) PasskeyVerifyBegin(req *types.VerifyBeginRequest) (resp *types.VerifyBeginResponse, err error) {
	// todo: add your logic here and delete this line

	return
}
