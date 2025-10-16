// UpdateOne 更新一条数据
// data 中主键字段必须有值，零值不会被更新
func ({{.firstTableChar}} *{{.upperTableName}}Repo) UpdateOne(ctx context.Context, newData *{{.dbName}}_model.{{.upperTableName}}) error {
	dao := {{.dbName}}_dao.Use({{.firstTableChar}}.db).{{.upperTableName}}
	_, err := dao.WithContext(ctx).Updates(newData)
	if err != nil {
		return err
	}
	return nil
}
{{- if .haveDeletedAt }}
// UpdateOneUnscoped 更新一条数据（包括软删除）
// data 中主键字段必须有值，零值不会被更新
func ({{.firstTableChar}} *{{.upperTableName}}Repo) UpdateOneUnscoped(ctx context.Context, newData *{{.dbName}}_model.{{.upperTableName}}) error {
	dao := {{.dbName}}_dao.Use({{.firstTableChar}}.db).{{.upperTableName}}
	_, err := dao.WithContext(ctx).Unscoped().Updates(newData)
	if err != nil {
		return err
	}
	return nil
}
{{- end }}