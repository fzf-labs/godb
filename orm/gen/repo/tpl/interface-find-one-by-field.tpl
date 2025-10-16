// FindOneBy{{.upperField}} 根据{{.lowerField}}查询一条数据
FindOneBy{{.upperField}}(ctx context.Context, {{.lowerField}} {{.dataType}}) (*{{.dbName}}_model.{{.upperTableName}}, error)
{{- if .haveDeletedAt }}
// FindOneUnscopedBy{{.upperField}} 根据{{.lowerField}}查询一条数据（包括软删除）
FindOneUnscopedBy{{.upperField}}(ctx context.Context, {{.lowerField}} {{.dataType}}) (*{{.dbName}}_model.{{.upperTableName}}, error)
{{- end }}