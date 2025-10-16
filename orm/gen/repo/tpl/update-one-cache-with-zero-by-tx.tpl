// UpdateOneCacheWithZeroByTx 更新一条数据(事务),包含零值，并删除缓存
// data 中主键字段必须有值,并且会更新所有字段,包括零值
// oldData 旧数据，删除缓存时使用
func ({{.firstTableChar}} *{{.upperTableName}}Repo) UpdateOneCacheWithZeroByTx(ctx context.Context, tx *{{.dbName}}_dao.Query, newData *{{.dbName}}_model.{{.upperTableName}}, oldData *{{.dbName}}_model.{{.upperTableName}}) error {
	dao := tx.{{.upperTableName}}
	_, err := dao.WithContext(ctx).Select(dao.ALL.WithTable("")).Updates(newData)
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
// UpdateOneUnscopedCacheWithZeroByTx 更新一条数据(事务),包含零值，并删除缓存（包括软删除）
// data 中主键字段必须有值,并且会更新所有字段,包括零值
// oldData 旧数据，删除缓存时使用
func ({{.firstTableChar}} *{{.upperTableName}}Repo) UpdateOneUnscopedCacheWithZeroByTx(ctx context.Context, tx *{{.dbName}}_dao.Query, newData *{{.dbName}}_model.{{.upperTableName}}, oldData *{{.dbName}}_model.{{.upperTableName}}) error {
	dao := tx.{{.upperTableName}}
	_, err := dao.WithContext(ctx).Unscoped().Select(dao.ALL.WithTable("")).Updates(newData)
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