// DeleteMultiCacheBy{{.upperFields}}Tx 根据{{.lowerField}}删除多条数据，并删除缓存(事务)
DeleteMultiCacheBy{{.upperFields}}Tx(ctx context.Context,tx *{{.dbName}}_dao.Query, {{.fieldAndDataTypes}}) error