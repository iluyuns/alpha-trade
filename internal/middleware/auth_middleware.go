package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/iluyuns/alpha-trade/internal/pkg/ctxval"
	"github.com/iluyuns/alpha-trade/internal/pkg/jwt"
	"github.com/iluyuns/alpha-trade/internal/pkg/revocation"
	"github.com/iluyuns/alpha-trade/internal/query"
	"github.com/zeromicro/go-zero/rest/httpx"
)

type AuthMiddleware struct {
	secret     string
	auditLogs  *query.AuditLogsCustom
	revocation revocation.RevocationManager
}

func NewAuthMiddleware(secret string, auditLogs *query.AuditLogsCustom, revocation revocation.RevocationManager) *AuthMiddleware {
	return &AuthMiddleware{
		secret:     secret,
		auditLogs:  auditLogs,
		revocation: revocation,
	}
}

func (m *AuthMiddleware) Handle(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			httpx.Error(w, jwt.ErrInvalidToken)
			return
		}

		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
		claims, err := jwt.ParseToken(m.secret, tokenStr)
		if err != nil {
			httpx.Error(w, err)
			return
		}

		// 增加撤销校验：检查签发时间是否在撤销点之前
		if m.revocation.IsRevoked(r.Context(), claims.UserId, claims.IssuedAt.Time) {
			httpx.Error(w, jwt.ErrInvalidToken)
			return
		}

		// 安全检查：IP 绑定校验 (防止会话劫持)
		currentIp := httpx.GetRemoteAddr(r)
		if claims.IssuedIp != "" && claims.IssuedIp != currentIp {
			// 记录会话被撤销的审计日志
			_ = m.auditLogs.RecordAction(r.Context(), claims.UserId, currentIp, "SESSION_REVOKED",
				"", "", fmt.Sprintf("{\"old\":\"%s\", \"new\":\"%s\"}", claims.IssuedIp, currentIp), false)
			httpx.Error(w, jwt.ErrInvalidToken)
			return
		}

		ctx := context.WithValue(r.Context(), ctxval.UIDKey, claims.UserId)
		ctx = context.WithValue(ctx, ctxval.ScopeKey, string(claims.Scope))
		ctx = context.WithValue(ctx, ctxval.TokenKey, tokenStr)
		if claims.ExpiresAt != nil {
			ctx = context.WithValue(ctx, ctxval.ExpKey, claims.ExpiresAt.Time)
		}

		next(w, r.WithContext(ctx))
	}
}
