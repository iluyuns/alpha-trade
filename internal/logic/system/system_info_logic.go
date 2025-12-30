package system

import (
	"context"

	"github.com/iluyuns/alpha-trade/internal/svc"
	"github.com/iluyuns/alpha-trade/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type SystemInfoLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSystemInfoLogic(ctx context.Context, svcCtx *svc.ServiceContext) SystemInfoLogic {
	return SystemInfoLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SystemInfoLogic) SystemInfo(req *types.SystemInfoReq) (resp *types.SystemInfoResp, err error) {
	// todo: add your logic here and delete this line

	return
}
