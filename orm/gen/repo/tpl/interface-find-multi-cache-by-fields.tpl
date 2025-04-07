// FindMultiCacheBy{{.upperFields}} 根据{{.upperFields}}查询多条数据，并设置缓存
FindMultiCacheBy{{.upperFields}}(ctx context.Context, {{.fieldAndDataTypes}}) ([]*{{.dbName}}_model.{{.upperTableName}}, error)