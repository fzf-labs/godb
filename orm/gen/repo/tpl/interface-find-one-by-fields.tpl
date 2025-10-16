// FindOneBy{{.upperFields}} 根据{{.upperFields}}查询一条数据
FindOneBy{{.upperFields}}(ctx context.Context, {{.fieldAndDataTypes}}) (*{{.dbName}}_model.{{.upperTableName}}, error)
{{- if .haveDeletedAt }}
// FindOneUnscopedBy{{.upperFields}} 根据{{.upperFields}}查询一条数据（包括软删除）
FindOneUnscopedBy{{.upperFields}}(ctx context.Context, {{.fieldAndDataTypes}}) (*{{.dbName}}_model.{{.upperTableName}}, error)
{{- end }}