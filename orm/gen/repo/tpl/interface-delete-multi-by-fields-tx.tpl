// DeleteMultiBy{{.upperFields}}Tx 根据{{.lowerField}}删除多条数据(事务)
DeleteMultiBy{{.upperFields}}Tx(ctx context.Context,tx *{{.dbName}}_dao.Query, {{.fieldAndDataTypes}}) error