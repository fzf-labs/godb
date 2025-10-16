// UpdateBatchBy{{.upperFields}}Tx 根据字段{{.upperFields}}批量更新(事务),零值会被更新
func ({{.firstTableChar}} *{{.upperTableName}}Repo) UpdateBatchBy{{.upperFields}}Tx(ctx context.Context,tx *{{.dbName}}_dao.Query, {{.fieldAndDataTypes}}, data map[string]interface{}) error {
	dao := tx.{{.upperTableName}}
	_, err := dao.WithContext(ctx).Where({{.whereFields}}).Updates(data)
	if err != nil {
		return err
	}
	return nil
}
{{- if .haveDeletedAt }}
// UpdateBatchUnscopedBy{{.upperFields}}Tx 根据字段{{.upperFields}}批量更新(事务),零值会被更新（包括软删除）
func ({{.firstTableChar}} *{{.upperTableName}}Repo) UpdateBatchUnscopedBy{{.upperFields}}Tx(ctx context.Context,tx *{{.dbName}}_dao.Query, {{.fieldAndDataTypes}}, data map[string]interface{}) error {
	dao := tx.{{.upperTableName}}
	_, err := dao.WithContext(ctx).Unscoped().Where({{.whereFields}}).Updates(data)
	if err != nil {
		return err
	}
	return nil
}
{{- end }}