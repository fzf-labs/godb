// UpdateOneCacheWithZero 更新一条数据,包含零值，并删除缓存
UpdateOneWithZero(ctx context.Context, newData *{{.dbName}}_model.{{.upperTableName}}) error
{{- if .haveDeletedAt }}
// UpdateOneUnscopedWithZero 更新一条数据,包含零值（包括软删除）
UpdateOneUnscopedWithZero(ctx context.Context, newData *{{.dbName}}_model.{{.upperTableName}}) error
{{- end }}