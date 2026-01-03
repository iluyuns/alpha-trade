package jwt

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

type TokenScope string

const (
	// ScopeBaseAuth 基础认证：已完成密码或 OAuth 登录，待 2FA 验证。
	ScopeBaseAuth TokenScope = "base_auth"

	// ScopeMfaAuth 完整认证：已完成 2FA 验证，具有常规业务权限。
	ScopeMfaAuth TokenScope = "mfa_auth"

	// ScopeSudoAuth 提级认证：用于高危操作（如提现、更改安全设置）。
	ScopeSudoAuth TokenScope = "sudo_auth"
)

// User Roles
const (
	RoleAdmin    = "ADMIN"
	RoleOperator = "OPERATOR"
	RoleViewer   = "VIEWER"
)

const (
	// ExpireBaseAuth 基础认证令牌有效期：5 分钟 (通常仅用于完成 MFA)
	ExpireBaseAuth = 300

	// ExpireMfaAuth 完整认证令牌有效期：2 小时 (标准会话)
	ExpireMfaAuth = 7200

	// ExpireSudoAuth 提级授权有效期：5 分钟
	ExpireSudoAuth = 300
)

var (
	ErrInvalidToken = errors.New("invalid token")
	ErrExpiredToken = errors.New("token expired")
)

type CustomClaims struct {
	UserId   int64      `json:"uid"`
	Scope    TokenScope `json:"scp"`
	IssuedIp string     `json:"iip,omitempty"` // 签发时的 IP 地址，用于防劫持校验
	jwt.RegisteredClaims
}

// GenerateToken 生成具有特定作用域和有效期的 Token
func GenerateToken(secret string, userId int64, scope TokenScope, expireSeconds int64) (string, error) {
	return GenerateTokenWithIp(secret, userId, scope, expireSeconds, "")
}

// GenerateTokenWithIp 生成绑定 IP 的 Token
func GenerateTokenWithIp(secret string, userId int64, scope TokenScope, expireSeconds int64, ip string) (string, error) {
	now := time.Now()
	claims := CustomClaims{
		UserId:   userId,
		Scope:    scope,
		IssuedIp: ip,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(time.Duration(expireSeconds) * time.Second)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

// ParseToken 解析并验证 Token
func ParseToken(secret string, tokenString string) (*CustomClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})

	if err != nil {
		if ve, ok := err.(*jwt.ValidationError); ok {
			if ve.Errors&jwt.ValidationErrorExpired != 0 {
				return nil, ErrExpiredToken
			}
		}
		return nil, ErrInvalidToken
	}

	if claims, ok := token.Claims.(*CustomClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, ErrInvalidToken
}
