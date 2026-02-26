package ctxval

import (
	"context"
	"time"
)

type contextKey string

const (
	IPKey    contextKey = "ip"
	UAKey    contextKey = "ua"
	UIDKey   contextKey = "uid"
	ScopeKey contextKey = "scope"
	TokenKey contextKey = "token"
	ExpKey   contextKey = "exp"
)

func GetIP(ctx context.Context) string {
	if v, ok := ctx.Value(IPKey).(string); ok {
		return v
	}
	return ""
}

func GetUA(ctx context.Context) string {
	if v, ok := ctx.Value(UAKey).(string); ok {
		return v
	}
	return ""
}

func GetUID(ctx context.Context) int64 {
	if v, ok := ctx.Value(UIDKey).(int64); ok {
		return v
	}
	return 0
}

func GetScope(ctx context.Context) string {
	if v, ok := ctx.Value(ScopeKey).(string); ok {
		return v
	}
	return ""
}

func GetToken(ctx context.Context) string {
	if v, ok := ctx.Value(TokenKey).(string); ok {
		return v
	}
	return ""
}

func GetExp(ctx context.Context) (time.Time, bool) {
	if v, ok := ctx.Value(ExpKey).(time.Time); ok {
		return v, true
	}
	return time.Time{}, false
}

