package query

// ExecutionsCustom is the custom extension for Executions.
// Add your custom methods here. This file will NOT be overwritten on regeneration.
type ExecutionsCustom struct {
	*executionsDo
}

// NewExecutions creates a new Executions data accessor with custom methods.
// Use this constructor to get both generated and custom methods.
func NewExecutions(db Executor) *ExecutionsCustom {
	return &ExecutionsCustom{
		executionsDo: executions.WithDB(db).(*executionsDo),
	}
}

// Example custom method (you can remove or modify this):
// func (c *ExecutionsCustom) FindByCustomCondition(ctx context.Context, condition string) ([]*Executions, error) {
// 	return c.Where(c.Field.ID.Gt(0)).Find(ctx)
// }
