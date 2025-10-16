// DeleteMultiBy{{.upperField}} 根据{{.upperField}}删除多条数据
DeleteMultiBy{{.upperField}}(ctx context.Context, {{.lowerField}} {{.dataType}}) error
{{- if .haveDeletedAt }}
// DeleteMultiUnscopedBy{{.upperField}} 根据{{.upperField}}删除多条数据
DeleteMultiUnscopedBy{{.upperField}}(ctx context.Context, {{.lowerField}} {{.dataType}}) error
{{- end }}