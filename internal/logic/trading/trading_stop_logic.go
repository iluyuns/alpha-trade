package trading

import (
	"context"

	"github.com/iluyuns/alpha-trade/internal/svc"
	"github.com/iluyuns/alpha-trade/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type TradingStopLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewTradingStopLogic(ctx context.Context, svcCtx *svc.ServiceContext) *TradingStopLogic {
	return &TradingStopLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *TradingStopLogic) TradingStop() (resp *types.TradingStopResponse, err error) {
	// 检查 TradingLoop 是否初始化
	if l.svcCtx.TradingLoop == nil {
		return &types.TradingStopResponse{
			Success: false,
			Message: "Trading components are not initialized",
		}, nil
	}

	// 检查是否已经停止
	if !l.svcCtx.TradingLoop.IsStarted() {
		return &types.TradingStopResponse{
			Success: true,
			Message: "Trading loop is already stopped",
		}, nil
	}

	// 停止交易循环
	l.svcCtx.TradingLoop.Stop()

	l.Infof("Trading loop stopped via API")
	return &types.TradingStopResponse{
		Success: true,
		Message: "Trading loop stopped successfully",
	}, nil
}
