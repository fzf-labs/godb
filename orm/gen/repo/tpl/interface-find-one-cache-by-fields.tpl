// FindOneCacheBy{{.upperFields}} 根据{{.upperFields}}查询一条数据，并设置缓存
FindOneCacheBy{{.upperFields}}(ctx context.Context, {{.fieldAndDataTypes}}) (*{{.dbName}}_model.{{.upperTableName}}, error)
{{- if .haveDeletedAt }}
// FindOneUnscopedCacheBy{{.upperFields}} 根据{{.upperFields}}查询一条数据（包括软删除），并设置缓存
FindOneUnscopedCacheBy{{.upperFields}}(ctx context.Context, {{.fieldAndDataTypes}}) (*{{.dbName}}_model.{{.upperTableName}}, error)
{{- end }}