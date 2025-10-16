// FindMultiCacheBy{{.upperFieldPlural}} 根据{{.lowerFieldPlural}}查询多条数据，并设置缓存
FindMultiCacheBy{{.upperFieldPlural}}(ctx context.Context, {{.lowerFieldPlural}} []{{.dataType}}) ([]*{{.dbName}}_model.{{.upperTableName}}, error)
{{- if .haveDeletedAt }}
// FindMultiUnscopedCacheBy{{.upperFieldPlural}} 根据{{.lowerFieldPlural}}查询多条数据（包括软删除），并设置缓存
FindMultiUnscopedCacheBy{{.upperFieldPlural}}(ctx context.Context, {{.lowerFieldPlural}} []{{.dataType}}) ([]*{{.dbName}}_model.{{.upperTableName}}, error)
{{- end }}