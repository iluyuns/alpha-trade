package query

import "context"

// UserAccessLogsCustom is the custom extension for UserAccessLogs.
// Add your custom methods here. This file will NOT be overwritten on regeneration.
type UserAccessLogsCustom struct {
	*userAccessLogsDo
}

// NewUserAccessLogs creates a new UserAccessLogs data accessor with custom methods.
// Use this constructor to get both generated and custom methods.
func NewUserAccessLogs(db Executor) *UserAccessLogsCustom {
	return &UserAccessLogsCustom{
		userAccessLogsDo: userAccessLogs.WithDB(db).(*userAccessLogsDo),
	}
}

// Insert creates a new user access log record (compatibility method)
func (c *UserAccessLogsCustom) Insert(ctx context.Context, log *UserAccessLogs) (*UserAccessLogs, error) {
	return c.Create(ctx, log)
}
