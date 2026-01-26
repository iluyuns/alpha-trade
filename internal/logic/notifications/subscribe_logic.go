package notifications

import (
	"context"
	"encoding/json"

	"github.com/iluyuns/alpha-trade/internal/svc"
	"github.com/iluyuns/alpha-trade/internal/types"
	"github.com/zeromicro/go-zero/core/logx"
)

type SubscribeLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSubscribeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SubscribeLogic {
	return &SubscribeLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SubscribeLogic) Subscribe(req *types.SubscribeReq) (resp *types.SubscribeResp, err error) {
	// 解析订阅信息
	var subscription map[string]interface{}
	if err := json.Unmarshal([]byte(req.Subscription), &subscription); err != nil {
		return &types.SubscribeResp{
			Success: false,
			Message: "无效的订阅信息",
		}, nil
	}

	// TODO: 存储订阅信息到数据库
	// 1. 获取当前用户 ID
	// 2. 存储 subscription JSON 到 push_subscriptions 表
	// 3. 关联用户 ID

	l.Infof("User subscribed to push notifications: %v", subscription)

	return &types.SubscribeResp{
		Success: true,
		Message: "订阅成功",
	}, nil
}
