// FindMultiCacheBy{{.upperField}} 根据{{.lowerField}}查询多条数据并设置缓存
FindMultiCacheBy{{.upperField}}(ctx context.Context, {{.lowerField}} {{.dataType}}) ([]*{{.dbName}}_model.{{.upperTableName}}, error)