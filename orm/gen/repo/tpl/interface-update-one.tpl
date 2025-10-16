// UpdateOne 更新一条数据
UpdateOne(ctx context.Context, newData *{{.dbName}}_model.{{.upperTableName}}) error
{{- if .haveDeletedAt }}
// UpdateOneUnscoped 更新一条数据（包括软删除）
UpdateOneUnscoped(ctx context.Context, newData *{{.dbName}}_model.{{.upperTableName}}) error
{{- end }}