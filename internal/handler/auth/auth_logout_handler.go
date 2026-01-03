package auth

import (
	"context"
	"net/http"

	"github.com/iluyuns/alpha-trade/internal/logic/auth"
	"github.com/iluyuns/alpha-trade/internal/pkg/ctxval"
	"github.com/iluyuns/alpha-trade/internal/svc"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func AuthLogoutHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 注入 IP 和 User-Agent 到 Context
		ctx := r.Context()
		ctx = context.WithValue(ctx, ctxval.IPKey, httpx.GetRemoteAddr(r))
		ctx = context.WithValue(ctx, ctxval.UAKey, r.Header.Get("User-Agent"))

		l := auth.NewAuthLogoutLogic(ctx, svcCtx)
		resp, err := l.AuthLogout()
		if err != nil {
			httpx.Error(w, err)
		} else {
			httpx.OkJson(w, resp)
		}
	}
}
