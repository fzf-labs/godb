// DeleteMultiBy{{.upperFieldPlural}} 根据{{.upperFieldPlural}}删除多条数据
DeleteMultiBy{{.upperFieldPlural}}(ctx context.Context, {{.lowerFieldPlural}} []{{.dataType}}) error
{{- if .haveDeletedAt }}
// DeleteMultiUnscopedBy{{.upperFieldPlural}} 根据{{.upperFieldPlural}}删除多条数据
DeleteMultiUnscopedBy{{.upperFieldPlural}}(ctx context.Context, {{.lowerFieldPlural}} []{{.dataType}}) error
{{- end }}