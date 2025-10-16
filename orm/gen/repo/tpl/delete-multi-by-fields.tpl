// DeleteMultiBy{{.upperFields}} 根据{{.lowerField}}删除多条数据
func ({{.firstTableChar}} *{{.upperTableName}}Repo) DeleteMultiBy{{.upperFields}}(ctx context.Context, {{.fieldAndDataTypes}}) error {
	dao := {{.dbName}}_dao.Use({{.firstTableChar}}.db).{{.upperTableName}}
	_, err := dao.WithContext(ctx).Where({{.whereFields}}).Delete()
	if err != nil {
		return err
	}
	return nil
}
{{- if .haveDeletedAt }}
// DeleteMultiUnscopedBy{{.upperFields}} 根据{{.lowerField}}删除多条数据
func ({{.firstTableChar}} *{{.upperTableName}}Repo) DeleteMultiUnscopedBy{{.upperFields}}(ctx context.Context, {{.fieldAndDataTypes}}) error {
	dao := {{.dbName}}_dao.Use({{.firstTableChar}}.db).{{.upperTableName}}
	_, err := dao.WithContext(ctx).Unscoped().Where({{.whereFields}}).Delete()
	if err != nil {
		return err
	}
	return nil
}
{{- end }}