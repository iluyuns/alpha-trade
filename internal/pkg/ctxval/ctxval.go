package ctxval

import "context"

type contextKey string

const (
	IPKey contextKey = "ip"
	UAKey contextKey = "ua"
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

