package main

import (
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

	fmt.Printf("Starting server at %s:%d...\n", c.Host, c.Port)
	server.Start()
}
