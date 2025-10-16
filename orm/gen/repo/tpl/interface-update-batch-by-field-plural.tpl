// UpdateBatchBy{{.upperFieldPlural}} 根据字段{{.upperFieldPlural}}批量更新,零值会被更新
UpdateBatchBy{{.upperFieldPlural}}(ctx context.Context, {{.lowerFieldPlural}} []{{.dataType}}, data map[string]interface{}) error
{{- if .haveDeletedAt }}
// UpdateBatchUnscopedBy{{.upperFieldPlural}} 根据字段{{.upperFieldPlural}}批量更新,零值会被更新（包括软删除）
UpdateBatchUnscopedBy{{.upperFieldPlural}}(ctx context.Context, {{.lowerFieldPlural}} []{{.dataType}}, data map[string]interface{}) error
{{- end }}