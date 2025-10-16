// FindMultiBy{{.upperFields}} 根据{{.upperFields}}查询多条数据，并设置缓存
FindMultiBy{{.upperFields}}(ctx context.Context, {{.fieldAndDataTypes}}) ([]*{{.dbName}}_model.{{.upperTableName}}, error)
{{- if .haveDeletedAt }}
// FindMultiUnscopedBy{{.upperFields}} 根据{{.upperFields}}查询多条数据（包括软删除），并设置缓存
FindMultiUnscopedBy{{.upperFields}}(ctx context.Context, {{.fieldAndDataTypes}}) ([]*{{.dbName}}_model.{{.upperTableName}}, error)
{{- end }}