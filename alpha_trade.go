package main

import (
	"context"
	"flag"
	"fmt"

	"github.com/iluyuns/alpha-trade/internal/config"
	"github.com/iluyuns/alpha-trade/internal/handler"
	"github.com/iluyuns/alpha-trade/internal/svc"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/proc"
	"github.com/zeromicro/go-zero/rest"
)

var configFile = flag.String("f", "etc/alpha_trade.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)

	ctx, err := svc.NewServiceContext(c)
	if err != nil {
		panic(err)
	}

	// 注册优雅退出回调
	proc.AddShutdownListener(func() {
		logx.Info("Shutting down service context...")
		if err := ctx.Close(); err != nil {
			logx.Errorf("Error during service context shutdown: %v", err)
		}
	})

	server := rest.MustNewServer(c.RestConf)
	defer server.Stop()

	handler.RegisterHandlers(server, ctx)

	// 如果启用交易且模式为 auto 或 hybrid，自动启动交易循环
	if c.Trading.Enabled && (c.Trading.Mode == "auto" || c.Trading.Mode == "hybrid") {
		if ctx.TradingLoop != nil {
			if err := ctx.TradingLoop.Start(context.Background()); err != nil {
				logx.Errorf("Failed to start trading loop: %v", err)
				// 不中断启动，允许 API 服务器继续运行
			} else {
				logx.Infof("Trading loop started automatically (mode: %s)", c.Trading.Mode)
			}
		} else {
			logx.Errorf("Trading is enabled but TradingLoop is not initialized. Check configuration.")
		}
	} else if c.Trading.Enabled {
		logx.Infof("Trading is enabled but mode is '%s', trading loop will not start automatically", c.Trading.Mode)
	}

	fmt.Printf("Starting server at %s:%d...\n", c.Host, c.Port)
	server.Start()
}
