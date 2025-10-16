// UpdateOneWithZeroByTx 更新一条数据(事务),包含零值，并删除缓存
UpdateOneWithZeroByTx(ctx context.Context, tx *{{.dbName}}_dao.Query, newData *{{.dbName}}_model.{{.upperTableName}}) error
{{- if .haveDeletedAt }}
// UpdateOneUnscopedWithZeroByTx 更新一条数据(事务),包含零值（包括软删除）
UpdateOneUnscopedWithZeroByTx(ctx context.Context, tx *{{.dbName}}_dao.Query, newData *{{.dbName}}_model.{{.upperTableName}}) error
{{- end }}