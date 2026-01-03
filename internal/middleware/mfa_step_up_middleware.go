package middleware

import (
	"net/http"

	"github.com/iluyuns/alpha-trade/internal/pkg/jwt"
	"github.com/zeromicro/go-zero/rest/httpx"
)

type MFAStepUpMiddleware struct {
	secret string
}

func NewMFAStepUpMiddleware(secret string) *MFAStepUpMiddleware {
	return &MFAStepUpMiddleware{
		secret: secret,
	}
}

func (m *MFAStepUpMiddleware) Handle(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 提级认证检查专有的 Sudo Header: X-Sudo-Token
		sudoToken := r.Header.Get("X-Sudo-Token")
		if sudoToken == "" {
			httpx.WriteJson(w, http.StatusForbidden, map[string]string{
				"error": "sudo_token_missing",
				"msg":   "High-risk operation requires sudo token",
			})
			return
		}

		claims, err := jwt.ParseToken(m.secret, sudoToken)
		if err != nil || claims.Scope != jwt.ScopeSudoAuth {
			httpx.WriteJson(w, http.StatusForbidden, map[string]string{
				"error": "sudo_invalid",
				"msg":   "Invalid or expired sudo token",
			})
			return
		}

		// 校验通过，确保 Sudo Token 属于当前登录用户
		if uid, ok := r.Context().Value("uid").(int64); ok {
			if uid != claims.UserId {
				httpx.WriteJson(w, http.StatusForbidden, map[string]string{"error": "sudo_mismatch"})
				return
			}
		}

		next(w, r)
	}
}
