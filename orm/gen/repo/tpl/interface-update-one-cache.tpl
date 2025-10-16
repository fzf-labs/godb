// UpdateOneCache 更新一条数据，并删除缓存
UpdateOneCache(ctx context.Context, newData *{{.dbName}}_model.{{.upperTableName}}, oldData *{{.dbName}}_model.{{.upperTableName}}) error
{{- if .haveDeletedAt }}
// UpdateOneUnscopedCache 更新一条数据，并删除缓存（包括软删除）
UpdateOneUnscopedCache(ctx context.Context, newData *{{.dbName}}_model.{{.upperTableName}}, oldData *{{.dbName}}_model.{{.upperTableName}}) error
{{- end }}