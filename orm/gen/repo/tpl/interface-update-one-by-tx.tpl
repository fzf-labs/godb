// UpdateOneByTx 更新一条数据(事务)
UpdateOneByTx(ctx context.Context, tx *{{.dbName}}_dao.Query, newData *{{.dbName}}_model.{{.upperTableName}}) error
{{- if .haveDeletedAt }}
// UpdateOneUnscopedByTx 更新一条数据(事务)（包括软删除）
UpdateOneUnscopedByTx(ctx context.Context, tx *{{.dbName}}_dao.Query, newData *{{.dbName}}_model.{{.upperTableName}}) error
{{- end }}