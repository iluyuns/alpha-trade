package query

// SettlementsCustom is the custom extension for Settlements.
// Add your custom methods here. This file will NOT be overwritten on regeneration.
type SettlementsCustom struct {
	*settlementsDo
}

// NewSettlements creates a new Settlements data accessor with custom methods.
// Use this constructor to get both generated and custom methods.
func NewSettlements(db Executor) *SettlementsCustom {
	return &SettlementsCustom{
		settlementsDo: settlements.WithDB(db).(*settlementsDo),
	}
}

// Example custom method (you can remove or modify this):
// func (c *SettlementsCustom) FindByCustomCondition(ctx context.Context, condition string) ([]*Settlements, error) {
// 	return c.Where(...).Find(ctx)
// }
