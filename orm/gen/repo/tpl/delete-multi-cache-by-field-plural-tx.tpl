// DeleteMultiCacheBy{{.upperFieldPlural}}Tx 根据{{.lowerFieldPlural}}删除多条数据，并删除缓存
func ({{.firstTableChar}} *{{.upperTableName}}Repo) DeleteMultiCacheBy{{.upperFieldPlural}}Tx(ctx context.Context,tx *{{.dbName}}_dao.Query, {{.lowerFieldPlural}} []{{.dataType}}) error {
	dao := tx.{{.upperTableName}}
	result, err := dao.WithContext(ctx).Where(dao.{{.upperField}}.In({{.lowerFieldPlural}}...)).Find()
	if err != nil {
		return err
	}
	if len(result) == 0 {
		return nil
	}
	_, err = dao.WithContext(ctx).Where(dao.{{.upperField}}.In({{.lowerFieldPlural}}...)).Delete()
	if err != nil {
		return err
	}
	err = {{.firstTableChar}}.DeleteIndexCache(ctx, result...)
	if err != nil {
		return err
	}
	return nil
}
