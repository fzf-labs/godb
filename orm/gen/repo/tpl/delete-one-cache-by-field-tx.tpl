// DeleteOneCacheBy{{.upperField}}Tx 根据{{.lowerField}}删除一条数据，并删除缓存
func ({{.firstTableChar}} *{{.upperTableName}}Repo) DeleteOneCacheBy{{.upperField}}Tx(ctx context.Context,tx *{{.dbName}}_dao.Query, {{.lowerField}} {{.dataType}}) error {
	dao := tx.{{.upperTableName}}
	result, err := dao.WithContext(ctx).Where(dao.{{.upperField}}.Eq({{.lowerField}})).First()
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	if result == nil {
		return nil
	}
	_, err = dao.WithContext(ctx).Where(dao.{{.upperField}}.Eq({{.lowerField}})).Delete()
	if err != nil {
		return err
	}
	err = {{.firstTableChar}}.DeleteIndexCache(ctx, result)
    if err != nil {
    	return err
    }
	return nil
}
{{- if .haveDeletedAt }}
// DeleteOneUnscopedCacheBy{{.upperField}}Tx 根据{{.lowerField}}删除一条数据，并删除缓存
func ({{.firstTableChar}} *{{.upperTableName}}Repo) DeleteOneUnscopedCacheBy{{.upperField}}Tx(ctx context.Context,tx *{{.dbName}}_dao.Query, {{.lowerField}} {{.dataType}}) error {
	dao := tx.{{.upperTableName}}
	result, err := dao.WithContext(ctx).Unscoped().Where(dao.{{.upperField}}.Eq({{.lowerField}})).First()
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	if result == nil {
		return nil
	}
	_, err = dao.WithContext(ctx).Unscoped().Where(dao.{{.upperField}}.Eq({{.lowerField}})).Delete()
	if err != nil {
		return err
	}
	err = {{.firstTableChar}}.DeleteIndexCache(ctx, result)
    if err != nil {
    	return err
    }
	return nil
}
{{- end }}