// FindMultiBy{{.upperFieldPlural}} 根据{{.lowerFieldPlural}}查询多条数据
FindMultiBy{{.upperFieldPlural}}(ctx context.Context, {{.lowerFieldPlural}} []{{.dataType}}) ([]*{{.dbName}}_model.{{.upperTableName}}, error)
{{- if .haveDeletedAt }}
// FindMultiUnscopedBy{{.upperFieldPlural}} 根据{{.lowerFieldPlural}}查询多条数据（包括软删除）
FindMultiUnscopedBy{{.upperFieldPlural}}(ctx context.Context, {{.lowerFieldPlural}} []{{.dataType}}) ([]*{{.dbName}}_model.{{.upperTableName}}, error)
{{- end }}