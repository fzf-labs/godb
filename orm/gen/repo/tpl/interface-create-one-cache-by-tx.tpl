// CreateOneCacheByTx 创建一条数据(事务), 并删除缓存
CreateOneCacheByTx(ctx context.Context, tx *{{.dbName}}_dao.Query, data *{{.dbName}}_model.{{.upperTableName}}) error