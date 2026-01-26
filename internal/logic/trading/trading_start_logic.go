package trading

import (
	"context"
	"fmt"

	"github.com/iluyuns/alpha-trade/internal/svc"
	"github.com/iluyuns/alpha-trade/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type TradingStartLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewTradingStartLogic(ctx context.Context, svcCtx *svc.ServiceContext) *TradingStartLogic {
	return &TradingStartLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *TradingStartLogic) TradingStart() (resp *types.TradingStartResponse, err error) {
	config := l.svcCtx.Config

	// 检查交易是否启用
	if !config.Trading.Enabled {
		return &types.TradingStartResponse{
			Success: false,
			Message: "Trading is not enabled. Please enable it in configuration first.",
		}, nil
	}

	// 检查交易模式
	if config.Trading.Mode == "auto" {
		return &types.TradingStartResponse{
			Success: false,
			Message: "Trading mode is 'auto', trading loop starts automatically. No need to start manually.",
		}, nil
	}

	// 检查 TradingLoop 是否初始化
	if l.svcCtx.TradingLoop == nil {
		return &types.TradingStartResponse{
			Success: false,
			Message: "Trading components are not initialized. Please check configuration.",
		}, nil
	}

	// 检查是否已经启动
	if l.svcCtx.TradingLoop.IsStarted() {
		return &types.TradingStartResponse{
			Success: true,
			Message: "Trading loop is already running",
		}, nil
	}

	// 启动交易循环
	if err := l.svcCtx.TradingLoop.Start(l.ctx); err != nil {
		l.Errorf("Failed to start trading loop: %v", err)
		return &types.TradingStartResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to start trading loop: %v", err),
		}, nil
	}

	l.Infof("Trading loop started via API")
	return &types.TradingStartResponse{
		Success: true,
		Message: "Trading loop started successfully",
	}, nil
}
