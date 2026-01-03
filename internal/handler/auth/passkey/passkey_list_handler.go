package passkey

import (
	"net/http"

	"github.com/iluyuns/alpha-trade/internal/logic/auth/passkey"
	"github.com/iluyuns/alpha-trade/internal/svc"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func PasskeyListHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := passkey.NewPasskeyListLogic(r.Context(), svcCtx)
		resp, err := l.PasskeyList()
		if err != nil {
			httpx.Error(w, err)
		} else {
			httpx.OkJson(w, resp)
		}
	}
}
