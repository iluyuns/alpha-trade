package passkey

import (
	"context"

	"github.com/iluyuns/alpha-trade/internal/svc"
	"github.com/iluyuns/alpha-trade/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type PasskeyRemoveLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewPasskeyRemoveLogic(ctx context.Context, svcCtx *svc.ServiceContext) PasskeyRemoveLogic {
	return PasskeyRemoveLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *PasskeyRemoveLogic) PasskeyRemove(req *types.RemoveRequest) (resp *types.AddFinishResponse, err error) {
	// todo: add your logic here and delete this line

	return
}
