package passkey

import (
	"net/http"

	"github.com/iluyuns/alpha-trade/internal/logic/auth/passkey"
	"github.com/iluyuns/alpha-trade/internal/svc"
	"github.com/iluyuns/alpha-trade/internal/types"

	"github.com/zeromicro/go-zero/rest/httpx"
)

func PasskeyVerifyBeginHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.VerifyBeginRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.Error(w, err)
			return
		}

		l := passkey.NewPasskeyVerifyBeginLogic(r.Context(), svcCtx)
		resp, err := l.PasskeyVerifyBegin(&req)
		if err != nil {
			httpx.Error(w, err)
		} else {
			httpx.OkJson(w, resp)
		}
	}
}
