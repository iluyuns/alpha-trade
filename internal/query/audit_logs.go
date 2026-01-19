package query

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

// Example custom method (you can remove or modify this):
// func (c *AuditLogsCustom) FindByCustomCondition(ctx context.Context, condition string) ([]*AuditLogs, error) {
// 	return c.Where(c.Field.ID.Gt(0)).Find(ctx)
// }
