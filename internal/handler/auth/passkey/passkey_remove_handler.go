package passkey

import (
	"net/http"

	"github.com/iluyuns/alpha-trade/internal/logic/auth/passkey"
	"github.com/iluyuns/alpha-trade/internal/svc"
	"github.com/iluyuns/alpha-trade/internal/types"

	"github.com/zeromicro/go-zero/rest/httpx"
)

func PasskeyRemoveHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.RemoveRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.Error(w, err)
			return
		}

		l := passkey.NewPasskeyRemoveLogic(r.Context(), svcCtx)
		resp, err := l.PasskeyRemove(&req)
		if err != nil {
			httpx.Error(w, err)
		} else {
			httpx.OkJson(w, resp)
		}
	}
}
