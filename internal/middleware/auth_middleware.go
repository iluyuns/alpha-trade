package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/iluyuns/alpha-trade/internal/model"
	"github.com/iluyuns/alpha-trade/internal/pkg/jwt"
	"github.com/iluyuns/alpha-trade/internal/pkg/revocation"
	"github.com/zeromicro/go-zero/rest/httpx"
)

type AuthMiddleware struct {
	secret     string
	model      model.UserAccessLogsModel
	revocation revocation.RevocationManager
}

func NewAuthMiddleware(secret string, model model.UserAccessLogsModel, revocation revocation.RevocationManager) *AuthMiddleware {
	return &AuthMiddleware{
		secret:     secret,
		model:      model,
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
			_, _ = m.model.Insert(r.Context(), &model.UserAccessLogs{
				UserId:    claims.UserId,
				IpAddress: currentIp,
				UserAgent: r.Header.Get("User-Agent"),
				Action:    "SESSION_REVOKED",
				Status:    "BLOCKED",
				Reason:    "IP_CHANGED",
				Details:   fmt.Sprintf("{\"old\":\"%s\", \"new\":\"%s\"}", claims.IssuedIp, currentIp),
			})
			httpx.Error(w, jwt.ErrInvalidToken)
			return
		}

		// 将 uid 注入 context，方便 logic 层通过 ctx.Value 获取
		ctx := context.WithValue(r.Context(), "uid", claims.UserId)
		// 注入 scope 供后续 logic 校验使用
		ctx = context.WithValue(ctx, "scope", claims.Scope)
		// 注入原始 token，供注销等逻辑使用
		ctx = context.WithValue(ctx, "token", tokenStr)
		// 注入过期时间，用于黑名单准确设置 TTL
		if claims.ExpiresAt != nil {
			ctx = context.WithValue(ctx, "exp", claims.ExpiresAt.Time)
		}

		next(w, r.WithContext(ctx))
	}
}
