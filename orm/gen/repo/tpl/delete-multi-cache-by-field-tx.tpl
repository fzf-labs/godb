// DeleteMultiCacheBy{{.upperField}}Tx 根据{{.lowerField}}删除多条数据，并删除缓存
func ({{.firstTableChar}} *{{.upperTableName}}Repo) DeleteMultiCacheBy{{.upperField}}Tx(ctx context.Context,tx *{{.dbName}}_dao.Query, {{.lowerField}} {{.dataType}}) error {
	dao := tx.{{.upperTableName}}
	result, err := dao.WithContext(ctx).Where(dao.{{.upperField}}.Eq({{.lowerField}})).Find()
	if err != nil {
		return err
	}
	if len(result) == 0 {
		return nil
	}
	_, err = dao.WithContext(ctx).Where(dao.{{.upperField}}.Eq({{.lowerField}})).Delete()
	if err != nil {
		return err
	}
	err = {{.firstTableChar}}.DeleteIndexCache(ctx, result...)
	if err != nil {
		return err
	}
	return nil
}
