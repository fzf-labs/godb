// UpdateBatchBy{{.upperFields}}Tx 根据主键{{.upperFields}}批量更新(事务),零值会被更新
UpdateBatchBy{{.upperFields}}Tx(ctx context.Context,tx *{{.dbName}}_dao.Query, {{.fieldAndDataTypes}}, data map[string]interface{}) error
{{- if .haveDeletedAt }}
// UpdateBatchUnscopedBy{{.upperFields}}Tx 根据主键{{.upperFields}}批量更新(事务),零值会被更新（包括软删除）
UpdateBatchUnscopedBy{{.upperFields}}Tx(ctx context.Context,tx *{{.dbName}}_dao.Query, {{.fieldAndDataTypes}}, data map[string]interface{}) error
{{- end }}