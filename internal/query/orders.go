package query

// OrdersCustom is the custom extension for Orders.
// Add your custom methods here. This file will NOT be overwritten on regeneration.
type OrdersCustom struct {
	*ordersDo
}

// NewOrders creates a new Orders data accessor with custom methods.
// Use this constructor to get both generated and custom methods.
func NewOrders(db Executor) *OrdersCustom {
	return &OrdersCustom{
		ordersDo: orders.WithDB(db).(*ordersDo),
	}
}

// Example custom method (you can remove or modify this):
// func (c *OrdersCustom) FindByCustomCondition(ctx context.Context, condition string) ([]*Orders, error) {
// 	return c.Where(c.Field.ID.Gt(0)).Find(ctx)
// }
