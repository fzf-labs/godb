// UpsertOneCacheByFields 根据fields字段Upsert一条数据, 并删除缓存
UpsertOneCacheByFields(ctx context.Context, data *{{.dbName}}_model.{{.upperTableName}},fields []string) error