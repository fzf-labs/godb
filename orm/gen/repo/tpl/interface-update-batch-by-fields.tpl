// UpdateBatchBy{{.upperFields}} 根据字段{{.upperFields}}批量更新,零值会被更新
UpdateBatchBy{{.upperFields}}(ctx context.Context,{{.fieldAndDataTypes}}, data map[string]interface{}) error
{{- if .haveDeletedAt }}
// UpdateBatchUnscopedBy{{.upperFields}} 根据字段{{.upperFields}}批量更新,零值会被更新（包括软删除）
UpdateBatchUnscopedBy{{.upperFields}}(ctx context.Context,{{.fieldAndDataTypes}}, data map[string]interface{}) error
{{- end }}