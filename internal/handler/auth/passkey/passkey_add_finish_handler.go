package passkey

import (
	"net/http"

	"github.com/iluyuns/alpha-trade/internal/logic/auth/passkey"
	"github.com/iluyuns/alpha-trade/internal/svc"
	"github.com/iluyuns/alpha-trade/internal/types"

	"github.com/zeromicro/go-zero/rest/httpx"
)

func PasskeyAddFinishHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.AddFinishRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.Error(w, err)
			return
		}

		l := passkey.NewPasskeyAddFinishLogic(r.Context(), svcCtx)
		resp, err := l.PasskeyAddFinish(&req)
		if err != nil {
			httpx.Error(w, err)
		} else {
			httpx.OkJson(w, resp)
		}
	}
}
