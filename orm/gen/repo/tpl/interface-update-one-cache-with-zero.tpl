// UpdateOneCacheWithZero 更新一条数据,包含零值，并删除缓存
UpdateOneCacheWithZero(ctx context.Context, newData *{{.dbName}}_model.{{.upperTableName}}, oldData *{{.dbName}}_model.{{.upperTableName}}) error
{{- if .haveDeletedAt }}
// UpdateOneUnscopedCacheWithZero 更新一条数据,包含零值，并删除缓存（包括软删除）
UpdateOneUnscopedCacheWithZero(ctx context.Context, newData *{{.dbName}}_model.{{.upperTableName}}, oldData *{{.dbName}}_model.{{.upperTableName}}) error
{{- end }}