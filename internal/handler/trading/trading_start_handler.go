package trading

import (
	"net/http"

	"github.com/iluyuns/alpha-trade/internal/logic/trading"
	"github.com/iluyuns/alpha-trade/internal/svc"

	"github.com/zeromicro/go-zero/rest/httpx"
)

func TradingStartHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := trading.NewTradingStartLogic(r.Context(), svcCtx)
		resp, err := l.TradingStart()
		if err != nil {
			httpx.Error(w, err)
		} else {
			httpx.OkJson(w, resp)
		}
	}
}
