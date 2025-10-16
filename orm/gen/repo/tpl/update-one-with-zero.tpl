// UpdateOneWithZero 更新一条数据,包含零值
// data 中主键字段必须有值,并且会更新所有字段,包括零值
func ({{.firstTableChar}} *{{.upperTableName}}Repo) UpdateOneWithZero(ctx context.Context, newData *{{.dbName}}_model.{{.upperTableName}}) error {
	dao := {{.dbName}}_dao.Use({{.firstTableChar}}.db).{{.upperTableName}}
	_, err := dao.WithContext(ctx).Select(dao.ALL.WithTable("")).Updates(newData)
	if err != nil {
		return err
	}
	return nil
}
{{- if .haveDeletedAt }}
// UpdateOneUnscopedWithZero 更新一条数据,包含零值（包括软删除）
// data 中主键字段必须有值,并且会更新所有字段,包括零值
func ({{.firstTableChar}} *{{.upperTableName}}Repo) UpdateOneUnscopedWithZero(ctx context.Context, newData *{{.dbName}}_model.{{.upperTableName}}) error {
	dao := {{.dbName}}_dao.Use({{.firstTableChar}}.db).{{.upperTableName}}
	_, err := dao.WithContext(ctx).Unscoped().Select(dao.ALL.WithTable("")).Updates(newData)
	if err != nil {
		return err
	}
	return nil
}
{{- end }}