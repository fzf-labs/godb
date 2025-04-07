// DeleteOneCacheBy{{.upperFields}}Tx 根据{{.upperFields}}删除一条数据，并删除缓存(事务)
DeleteOneCacheBy{{.upperFields}}Tx(ctx context.Context,tx *{{.dbName}}_dao.Query, {{.fieldAndDataTypes}}) error