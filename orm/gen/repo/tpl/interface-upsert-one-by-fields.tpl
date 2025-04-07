// UpsertOneByFields 根据fields字段Upsert一条数据
UpsertOneByFields(ctx context.Context, data *{{.dbName}}_model.{{.upperTableName}},fields []string) error