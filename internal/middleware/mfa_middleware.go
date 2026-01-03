package middleware

import (
	"net/http"

	"github.com/iluyuns/alpha-trade/internal/pkg/jwt"
	"github.com/zeromicro/go-zero/rest/httpx"
)

type MFAMiddleware struct {
}

func NewMFAMiddleware() *MFAMiddleware {
	return &MFAMiddleware{}
}

func (m *MFAMiddleware) Handle(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 基础认证校验：确保用户已完成登录后的 MFA 验证
		// 此时 AuthMiddleware 已经将 scope 注入 context
		scope, ok := r.Context().Value("scope").(jwt.TokenScope)
		if !ok || scope == jwt.ScopeBaseAuth {
			httpx.WriteJson(w, http.StatusForbidden, map[string]string{
				"error": "mfa_required",
				"msg":   "Please complete 2FA verification first",
			})
			return
		}

		next(w, r)
	}
}
