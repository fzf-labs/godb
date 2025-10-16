// UpdateOneCacheByTx 更新一条数据(事务)，并删除缓存
UpdateOneCacheByTx(ctx context.Context, tx *{{.dbName}}_dao.Query, newData *{{.dbName}}_model.{{.upperTableName}}, oldData *{{.dbName}}_model.{{.upperTableName}}) error
{{- if .haveDeletedAt }}
// UpdateOneUnscopedCacheByTx 更新一条数据(事务)，并删除缓存（包括软删除）
UpdateOneUnscopedCacheByTx(ctx context.Context, tx *{{.dbName}}_dao.Query, newData *{{.dbName}}_model.{{.upperTableName}}, oldData *{{.dbName}}_model.{{.upperTableName}}) error
{{- end }}