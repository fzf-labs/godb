// UpdateBatchBy{{.upperFields}} 根据字段{{.upperFields}}批量更新,零值会被更新
func ({{.firstTableChar}} *{{.upperTableName}}Repo) UpdateBatchBy{{.upperFields}}(ctx context.Context,{{.fieldAndDataTypes}}, data map[string]interface{}) error {
	dao := {{.dbName}}_dao.Use({{.firstTableChar}}.db).{{.upperTableName}}
	_, err := dao.WithContext(ctx).Where({{.whereFields}}).Updates(data)
	if err != nil {
		return err
	}
	return nil
}
{{- if .haveDeletedAt }}
// UpdateBatchUnscopedBy{{.upperFields}} 根据字段{{.upperFields}}批量更新,零值会被更新（包括软删除）
func ({{.firstTableChar}} *{{.upperTableName}}Repo) UpdateBatchUnscopedBy{{.upperFields}}(ctx context.Context,{{.fieldAndDataTypes}}, data map[string]interface{}) error {
	dao := {{.dbName}}_dao.Use({{.firstTableChar}}.db).{{.upperTableName}}
	_, err := dao.WithContext(ctx).Unscoped().Where({{.whereFields}}).Updates(data)
	if err != nil {
		return err
	}
	return nil
}
{{- end }}