// CreateOneCacheByTx 创建一条数据(事务), 并删除缓存
func ({{.firstTableChar}} *{{.upperTableName}}Repo) CreateOneCacheByTx(ctx context.Context, tx *{{.dbName}}_dao.Query, data *{{.dbName}}_model.{{.upperTableName}}) error {
	dao := tx.{{.upperTableName}}
	err := dao.WithContext(ctx).Create(data)
	if err != nil {
		return err
	}
	err = {{.firstTableChar}}.DeleteIndexCache(ctx,data)
    if err != nil {
    	return err
    }
	return nil
}