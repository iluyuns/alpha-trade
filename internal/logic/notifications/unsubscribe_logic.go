package notifications

import (
	"context"

	"github.com/iluyuns/alpha-trade/internal/svc"
	"github.com/iluyuns/alpha-trade/internal/types"
	"github.com/zeromicro/go-zero/core/logx"
)

type UnsubscribeLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUnsubscribeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UnsubscribeLogic {
	return &UnsubscribeLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UnsubscribeLogic) Unsubscribe(req *types.UnsubscribeReq) (resp *types.UnsubscribeResp, err error) {
	// TODO: 从数据库删除订阅信息
	// 1. 获取当前用户 ID
	// 2. 删除对应的订阅记录

	l.Infof("User unsubscribed from push notifications")

	return &types.UnsubscribeResp{
		Success: true,
	}, nil
}
