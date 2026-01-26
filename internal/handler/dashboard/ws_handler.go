package dashboard

import (
	"context"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"github.com/iluyuns/alpha-trade/internal/logic/dashboard"
	"github.com/iluyuns/alpha-trade/internal/svc"
	"github.com/iluyuns/alpha-trade/internal/types"
	"github.com/zeromicro/go-zero/core/logx"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // 允许跨域，生产环境应该检查 Origin
	},
}

func DashboardWSHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 升级为 WebSocket 连接
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			logx.Errorf("Failed to upgrade websocket: %v", err)
			return
		}
		defer conn.Close()

		// 创建 Dashboard Logic
		ctx := r.Context()
		logic := dashboard.NewDashboardLogic(ctx, svcCtx)

		// 启动推送 goroutine
		go pushDashboardData(conn, logic, ctx)

		// 保持连接，处理 ping/pong
		for {
			_, _, err := conn.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					logx.Errorf("WebSocket error: %v", err)
				}
				break
			}
		}
	}
}

func pushDashboardData(conn *websocket.Conn, logic *dashboard.DashboardLogic, ctx context.Context) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	// 立即发送一次数据
	sendDashboardData(conn, logic)

	for {
		select {
		case <-ticker.C:
			if err := sendDashboardData(conn, logic); err != nil {
				logx.Errorf("Failed to send dashboard data: %v", err)
				return
			}
		case <-ctx.Done():
			return
		}
	}
}

func sendDashboardData(conn *websocket.Conn, logic *dashboard.DashboardLogic) error {
	// 获取 Dashboard 数据
	resp, err := logic.Dashboard(&types.DashboardReq{})
	if err != nil {
		return err
	}

	// 构建 WebSocket 消息
	message := map[string]interface{}{
		"type":      "dashboard",
		"data":      resp,
		"timestamp": time.Now().Unix(),
	}

	// 发送 JSON 消息
	return conn.WriteJSON(message)
}
