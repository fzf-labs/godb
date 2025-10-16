// FindMultiCacheBy{{.upperField}} 根据{{.lowerField}}查询多条数据并设置缓存
FindMultiCacheBy{{.upperField}}(ctx context.Context, {{.lowerField}} {{.dataType}}) ([]*{{.dbName}}_model.{{.upperTableName}}, error)
{{- if .haveDeletedAt }}
// FindMultiUnscopedCacheBy{{.upperField}} 根据{{.lowerField}}查询多条数据（包括软删除）并设置缓存
FindMultiUnscopedCacheBy{{.upperField}}(ctx context.Context, {{.lowerField}} {{.dataType}}) ([]*{{.dbName}}_model.{{.upperTableName}}, error)
{{- end }}