package auth

import (
	"context"
	"net/http"

	"github.com/iluyuns/alpha-trade/internal/logic/auth"
	"github.com/iluyuns/alpha-trade/internal/pkg/ctxval"
	"github.com/iluyuns/alpha-trade/internal/svc"
	"github.com/iluyuns/alpha-trade/internal/types"

	"github.com/zeromicro/go-zero/rest/httpx"
)

func AuthLoginHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.LoginRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.Error(w, err)
			return
		}

		// 注入 IP 和 User-Agent 到 Context
		ctx := r.Context()
		ctx = context.WithValue(ctx, ctxval.IPKey, httpx.GetRemoteAddr(r))
		ctx = context.WithValue(ctx, ctxval.UAKey, r.Header.Get("User-Agent"))

		l := auth.NewAuthLoginLogic(ctx, svcCtx)
		resp, err := l.AuthLogin(&req)
		if err != nil {
			httpx.Error(w, err)
		} else {
			httpx.OkJson(w, resp)
		}
	}
}
