// UpdateBatchBy{{.upperFieldPlural}}Tx 根据字段{{.upperFieldPlural}}批量更新(事务),零值会被更新
UpdateBatchBy{{.upperFieldPlural}}Tx(ctx context.Context,tx *{{.dbName}}_dao.Query, {{.lowerFieldPlural}} []{{.dataType}}, data map[string]interface{}) error
{{- if .haveDeletedAt }}
// UpdateBatchUnscopedBy{{.upperFieldPlural}}Tx 根据字段{{.upperFieldPlural}}批量更新(事务),零值会被更新（包括软删除）
UpdateBatchUnscopedBy{{.upperFieldPlural}}Tx(ctx context.Context,tx *{{.dbName}}_dao.Query, {{.lowerFieldPlural}} []{{.dataType}}, data map[string]interface{}) error
{{- end }}