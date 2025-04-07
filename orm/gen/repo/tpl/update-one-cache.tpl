// UpdateOneCache 更新一条数据，并删除缓存
// data 中主键字段必须有值，零值不会被更新
// oldData 旧数据，删除缓存时使用
func ({{.firstTableChar}} *{{.upperTableName}}Repo) UpdateOneCache(ctx context.Context, newData *{{.dbName}}_model.{{.upperTableName}}, oldData *{{.dbName}}_model.{{.upperTableName}}) error {
	dao := {{.dbName}}_dao.Use({{.firstTableChar}}.db).{{.upperTableName}}
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