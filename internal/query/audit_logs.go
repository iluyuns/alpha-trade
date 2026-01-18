package query

import "context"

// AuditLogsCustom is the custom extension for AuditLogs.
// Add your custom methods here. This file will NOT be overwritten on regeneration.
type AuditLogsCustom struct {
	*auditLogsDo
}

// NewAuditLogs creates a new AuditLogs data accessor with custom methods.
// Use this constructor to get both generated and custom methods.
func NewAuditLogs(db Executor) *AuditLogsCustom {
	return &AuditLogsCustom{
		auditLogsDo: auditLogs.WithDB(db).(*auditLogsDo),
	}
}

// RecordAction 记录审计日志
func (c *AuditLogsCustom) RecordAction(ctx context.Context, userID int64, ip, action, targetType, targetID, changes string, isVerified bool) error {
	_, err := c.Create(ctx, &AuditLogs{
		UserID:     userID,
		IpAddress:  ip,
		Action:     action,
		TargetType: targetType,
		TargetID:   targetID,
		Changes:    changes,
		IsVerified: isVerified,
	})
	return err
}
