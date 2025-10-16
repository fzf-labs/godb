// UpdateOneCacheWithZeroByTx 更新一条数据(事务),包含零值，并删除缓存
UpdateOneCacheWithZeroByTx(ctx context.Context, tx *{{.dbName}}_dao.Query, newData *{{.dbName}}_model.{{.upperTableName}}, oldData *{{.dbName}}_model.{{.upperTableName}}) error
{{- if .haveDeletedAt }}
// UpdateOneUnscopedCacheWithZeroByTx 更新一条数据(事务),包含零值，并删除缓存（包括软删除）
UpdateOneUnscopedCacheWithZeroByTx(ctx context.Context, tx *{{.dbName}}_dao.Query, newData *{{.dbName}}_model.{{.upperTableName}}, oldData *{{.dbName}}_model.{{.upperTableName}}) error
{{- end }}