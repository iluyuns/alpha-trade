package query

// AssetSnapshotsCustom is the custom extension for AssetSnapshots.
// Add your custom methods here. This file will NOT be overwritten on regeneration.
type AssetSnapshotsCustom struct {
	*assetSnapshotsDo
}

// NewAssetSnapshots creates a new AssetSnapshots data accessor with custom methods.
// Use this constructor to get both generated and custom methods.
func NewAssetSnapshots(db Executor) *AssetSnapshotsCustom {
	return &AssetSnapshotsCustom{
		assetSnapshotsDo: assetSnapshots.WithDB(db).(*assetSnapshotsDo),
	}
}

// Example custom method (you can remove or modify this):
// func (c *AssetSnapshotsCustom) FindByCustomCondition(ctx context.Context, condition string) ([]*AssetSnapshots, error) {
// 	return c.Where(...).Find(ctx)
// }
