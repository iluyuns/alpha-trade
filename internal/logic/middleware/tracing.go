package middleware

import (
	"context"
)

// TraceInterceptor 链路追踪拦截器 (示例)
func TraceInterceptor(ctx context.Context, name string, next func(ctx context.Context) error) error {
	// 1. 从 context 或消息 Header 中提取 TraceID
	// 2. 开始新的 Span
	// 3. 执行业务逻辑
	// 4. 结束 Span 并记录错误（如果有）
	return next(ctx)
}

