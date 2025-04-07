// DeleteOneBy{{.upperField}}Tx 根据{{.lowerField}}删除一条数据(事务)
DeleteOneBy{{.upperField}}Tx(ctx context.Context,tx *{{.dbName}}_dao.Query, {{.lowerField}} {{.dataType}}) error