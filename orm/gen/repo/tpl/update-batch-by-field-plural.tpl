// UpdateBatchBy{{.upperFieldPlural}} 根据字段{{.upperFieldPlural}}批量更新,零值会被更新
func ({{.firstTableChar}} *{{.upperTableName}}Repo) UpdateBatchBy{{.upperFieldPlural}}(ctx context.Context, {{.lowerFieldPlural}} []{{.dataType}}, data map[string]interface{}) error {
	dao := {{.dbName}}_dao.Use({{.firstTableChar}}.db).{{.upperTableName}}
	_, err := dao.WithContext(ctx).Where(dao.{{.upperField}}.In({{.lowerFieldPlural}}...)).Updates(data)
	if err != nil {
		return err
	}
	return nil
}
{{- if .haveDeletedAt }}
// UpdateBatchUnscopedBy{{.upperFieldPlural}} 根据字段{{.upperFieldPlural}}批量更新,零值会被更新（包括软删除）
func ({{.firstTableChar}} *{{.upperTableName}}Repo) UpdateBatchUnscopedBy{{.upperFieldPlural}}(ctx context.Context, {{.lowerFieldPlural}} []{{.dataType}}, data map[string]interface{}) error {
	dao := {{.dbName}}_dao.Use({{.firstTableChar}}.db).{{.upperTableName}}
	_, err := dao.WithContext(ctx).Unscoped().Where(dao.{{.upperField}}.In({{.lowerFieldPlural}}...)).Updates(data)
	if err != nil {
		return err
	}
	return nil
}
{{- end }}