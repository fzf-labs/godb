// DeleteOneBy{{.upperField}} 根据{{.lowerField}}删除一条数据
DeleteOneBy{{.upperField}}(ctx context.Context, {{.lowerField}} {{.dataType}}) error
{{- if .haveDeletedAt }}
// DeleteOneUnscopedBy{{.upperField}} 根据{{.lowerField}}删除一条数据
DeleteOneUnscopedBy{{.upperField}}(ctx context.Context, {{.lowerField}} {{.dataType}}) error
{{- end }}