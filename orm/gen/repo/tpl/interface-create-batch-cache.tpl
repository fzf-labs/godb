// CreateBatchCache 批量创建数据, 并删除缓存
CreateBatchCache(ctx context.Context, data []*{{.dbName}}_model.{{.upperTableName}}, batchSize int) error