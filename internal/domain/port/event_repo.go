package port

import (
	"context"
	"time"
)

// MacroEvent 宏观事件/新闻
type MacroEvent struct {
	ID          string
	Type        string // "NEWS", "ANNOUNCEMENT", "REGULATORY"
	Title       string
	Content     string
	Severity    string    // "LOW", "MEDIUM", "HIGH", "CRITICAL"
	Source      string    // 数据源
	PublishTime time.Time // 发布时间
	ExpiryTime  time.Time // 失效时间（冷却窗口结束）
}

// EventRepo 宏观事件/新闻接口
// 用于风控：重大事件发生时触发交易冷却或降杠杆
type EventRepo interface {
	// GetActiveEvents 获取当前活跃事件
	// 返回在冷却窗口内的所有事件
	GetActiveEvents(ctx context.Context) ([]*MacroEvent, error)

	// GetEventsBySeverity 按严重程度过滤事件
	GetEventsBySeverity(ctx context.Context, severity string) ([]*MacroEvent, error)

	// SubscribeEvents 订阅事件流（实盘用）
	SubscribeEvents(ctx context.Context) (<-chan *MacroEvent, error)

	// IsInCooldown 检查是否在冷却期
	// 根据当前活跃的高危事件判断
	IsInCooldown(ctx context.Context) (bool, error)
}
