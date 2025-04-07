// DeleteMultiCacheBy{{.upperField}}Tx 根据{{.lowerField}}删除多条数据，并删除缓存
DeleteMultiCacheBy{{.upperField}}Tx(ctx context.Context,tx *{{.dbName}}_dao.Query, {{.lowerField}} {{.dataType}}) error