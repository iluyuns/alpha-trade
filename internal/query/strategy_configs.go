package query

// StrategyConfigsCustom is the custom extension for StrategyConfigs.
// Add your custom methods here. This file will NOT be overwritten on regeneration.
type StrategyConfigsCustom struct {
	*strategyConfigsDo
}

// NewStrategyConfigs creates a new StrategyConfigs data accessor with custom methods.
// Use this constructor to get both generated and custom methods.
func NewStrategyConfigs(db Executor) *StrategyConfigsCustom {
	return &StrategyConfigsCustom{
		strategyConfigsDo: strategyConfigs.WithDB(db).(*strategyConfigsDo),
	}
}

// Example custom method (you can remove or modify this):
// func (c *StrategyConfigsCustom) FindByCustomCondition(ctx context.Context, condition string) ([]*StrategyConfigs, error) {
// 	return c.Where(...).Find(ctx)
// }
