// UpsertOneByFieldsTx 根据fields字段Upsert一条数据(事务)
UpsertOneByFieldsTx(ctx context.Context,tx *{{.dbName}}_dao.Query, data *{{.dbName}}_model.{{.upperTableName}},fields []string) error