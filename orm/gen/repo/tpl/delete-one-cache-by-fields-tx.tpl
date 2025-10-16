// DeleteOneCacheBy{{.upperFields}}Tx 根据{{.lowerField}}删除一条数据，并删除缓存
func ({{.firstTableChar}} *{{.upperTableName}}Repo) DeleteOneCacheBy{{.upperFields}}Tx(ctx context.Context,tx *{{.dbName}}_dao.Query, {{.fieldAndDataTypes}}) error {
	dao := tx.{{.upperTableName}}
	result, err := dao.WithContext(ctx).Where({{.whereFields}}).First()
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	if result == nil {
		return nil
	}
	_, err = dao.WithContext(ctx).Where({{.whereFields}}).Delete()
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
// DeleteOneUnscopedCacheBy{{.upperFields}}Tx 根据{{.lowerField}}删除一条数据，并删除缓存
func ({{.firstTableChar}} *{{.upperTableName}}Repo) DeleteOneUnscopedCacheBy{{.upperFields}}Tx(ctx context.Context,tx *{{.dbName}}_dao.Query, {{.fieldAndDataTypes}}) error {
	dao := tx.{{.upperTableName}}
	result, err := dao.WithContext(ctx).Unscoped().Where({{.whereFields}}).First()
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	if result == nil {
		return nil
	}
	_, err = dao.WithContext(ctx).Unscoped().Where({{.whereFields}}).Delete()
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