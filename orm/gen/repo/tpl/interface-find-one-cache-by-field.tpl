// FindOneCacheBy{{.upperField}} 根据{{.lowerField}}查询一条数据，并设置缓存
FindOneCacheBy{{.upperField}}(ctx context.Context, {{.lowerField}} {{.dataType}}) (*{{.dbName}}_model.{{.upperTableName}}, error)
{{- if .haveDeletedAt }}
// FindOneUnscopedCacheBy{{.upperField}} 根据{{.lowerField}}查询一条数据（包括软删除），并设置缓存
FindOneUnscopedCacheBy{{.upperField}}(ctx context.Context, {{.lowerField}} {{.dataType}}) (*{{.dbName}}_model.{{.upperTableName}}, error)
{{- end }}