package query

// RiskRecordsCustom is the custom extension for RiskRecords.
// Add your custom methods here. This file will NOT be overwritten on regeneration.
type RiskRecordsCustom struct {
	*riskRecordsDo
}

// NewRiskRecords creates a new RiskRecords data accessor with custom methods.
// Use this constructor to get both generated and custom methods.
func NewRiskRecords(db Executor) *RiskRecordsCustom {
	return &RiskRecordsCustom{
		riskRecordsDo: riskRecords.WithDB(db).(*riskRecordsDo),
	}
}

// Example custom method (you can remove or modify this):
// func (c *RiskRecordsCustom) FindByCustomCondition(ctx context.Context, condition string) ([]*RiskRecords, error) {
// 	return c.Where(...).Find(ctx)
// }
