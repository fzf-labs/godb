// UpdateOneByTx 更新一条数据(事务)
UpdateOneByTx(ctx context.Context, tx *{{.dbName}}_dao.Query, newData *{{.dbName}}_model.{{.upperTableName}}) error