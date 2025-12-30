package passkey

import (
	"net/http"

	"github.com/iluyuns/alpha-trade/internal/logic/passkey"
	"github.com/iluyuns/alpha-trade/internal/svc"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func ListHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := passkey.NewListLogic(r.Context(), svcCtx)
		resp, err := l.List()
		if err != nil {
			httpx.Error(w, err)
		} else {
			httpx.OkJson(w, resp)
		}
	}
}
