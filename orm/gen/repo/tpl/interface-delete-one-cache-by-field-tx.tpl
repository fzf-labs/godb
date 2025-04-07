// DeleteOneCacheBy{{.upperField}}Tx 根据{{.lowerField}}删除一条数据，并删除缓存(事务)
DeleteOneCacheBy{{.upperField}}Tx(ctx context.Context,tx *{{.dbName}}_dao.Query, {{.lowerField}} {{.dataType}}) error