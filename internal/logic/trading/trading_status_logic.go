package trading

import (
	"context"

	"github.com/iluyuns/alpha-trade/internal/svc"
	"github.com/iluyuns/alpha-trade/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type TradingStatusLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewTradingStatusLogic(ctx context.Context, svcCtx *svc.ServiceContext) *TradingStatusLogic {
	return &TradingStatusLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *TradingStatusLogic) TradingStatus() (resp *types.TradingStatusResponse, err error) {
	config := l.svcCtx.Config

	resp = &types.TradingStatusResponse{
		Enabled:  config.Trading.Enabled,
		Mode:     config.Trading.Mode,
		Symbols:  config.Trading.Symbols,
		Interval: config.Trading.KlineInterval,
		Strategy: config.Trading.StrategyType,
	}

	// 检查交易循环状态
	if l.svcCtx.TradingLoop != nil {
		resp.Started = l.svcCtx.TradingLoop.IsStarted()
		if resp.Started {
			resp.Message = "Trading loop is running"
		} else {
			resp.Message = "Trading loop is stopped"
		}
	} else {
		resp.Started = false
		if !config.Trading.Enabled {
			resp.Message = "Trading is not enabled"
		} else {
			resp.Message = "Trading components not initialized"
		}
	}

	return resp, nil
}
