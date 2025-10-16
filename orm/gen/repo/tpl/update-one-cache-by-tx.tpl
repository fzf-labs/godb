// UpdateOneCacheByTx 更新一条数据(事务)，并删除缓存
// data 中主键字段必须有值，零值不会被更新
// oldData 旧数据，删除缓存时使用
func ({{.firstTableChar}} *{{.upperTableName}}Repo) UpdateOneCacheByTx(ctx context.Context, tx *{{.dbName}}_dao.Query, newData *{{.dbName}}_model.{{.upperTableName}}, oldData *{{.dbName}}_model.{{.upperTableName}}) error {
	dao := tx.{{.upperTableName}}
	_, err := dao.WithContext(ctx).Updates(newData)
	if err != nil {
		return err
	}
    err = {{.firstTableChar}}.DeleteIndexCache(ctx,oldData,newData)
    if err != nil {
        return err
    }
	return nil
}
{{- if .haveDeletedAt }}
// UpdateOneUnscopedCacheByTx 更新一条数据(事务)，并删除缓存（包括软删除）
// data 中主键字段必须有值，零值不会被更新
// oldData 旧数据，删除缓存时使用
func ({{.firstTableChar}} *{{.upperTableName}}Repo) UpdateOneUnscopedCacheByTx(ctx context.Context, tx *{{.dbName}}_dao.Query, newData *{{.dbName}}_model.{{.upperTableName}}, oldData *{{.dbName}}_model.{{.upperTableName}}) error {
	dao := tx.{{.upperTableName}}
	_, err := dao.WithContext(ctx).Unscoped().Updates(newData)
	if err != nil {
		return err
	}
    err = {{.firstTableChar}}.DeleteIndexCache(ctx,oldData,newData)
    if err != nil {
        return err
    }
	return nil
}
{{- end }}