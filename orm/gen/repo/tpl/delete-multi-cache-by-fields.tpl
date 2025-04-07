// DeleteMultiCacheBy{{.upperFields}} 根据{{.lowerField}}删除多条数据，并删除缓存
func ({{.firstTableChar}} *{{.upperTableName}}Repo) DeleteMultiCacheBy{{.upperFields}}(ctx context.Context, {{.fieldAndDataTypes}}) error {
	dao := {{.dbName}}_dao.Use({{.firstTableChar}}.db).{{.upperTableName}}
	result, err := dao.WithContext(ctx).Where({{.whereFields}}).Find()
	if err != nil {
		return err
	}
	if len(result) == 0 {
		return nil
	}
	_, err = dao.WithContext(ctx).Where({{.whereFields}}).Delete()
	if err != nil {
		return err
	}
	err = {{.firstTableChar}}.DeleteIndexCache(ctx, result...)
	if err != nil {
		return err
	}
	return nil
}