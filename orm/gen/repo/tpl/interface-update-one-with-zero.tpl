// UpdateOneCacheWithZero 更新一条数据,包含零值，并删除缓存
UpdateOneWithZero(ctx context.Context, newData *{{.dbName}}_model.{{.upperTableName}}) error