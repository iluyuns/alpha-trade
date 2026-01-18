package query

// ExchangeAccountsCustom is the custom extension for ExchangeAccounts.
// Add your custom methods here. This file will NOT be overwritten on regeneration.
type ExchangeAccountsCustom struct {
	*exchangeAccountsDo
}

// NewExchangeAccounts creates a new ExchangeAccounts data accessor with custom methods.
// Use this constructor to get both generated and custom methods.
func NewExchangeAccounts(db Executor) *ExchangeAccountsCustom {
	return &ExchangeAccountsCustom{
		exchangeAccountsDo: exchangeAccounts.WithDB(db).(*exchangeAccountsDo),
	}
}

// Example custom method (you can remove or modify this):
// func (c *ExchangeAccountsCustom) FindByCustomCondition(ctx context.Context, condition string) ([]*ExchangeAccounts, error) {
// 	return c.Where(...).Find(ctx)
// }
