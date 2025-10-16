// DeleteMultiBy{{.upperFields}} 根据{{.lowerField}}删除多条数据
DeleteMultiBy{{.upperFields}}(ctx context.Context, {{.fieldAndDataTypes}}) error
{{- if .haveDeletedAt }}
// DeleteMultiUnscopedBy{{.upperFields}} 根据{{.lowerField}}删除多条数据
DeleteMultiUnscopedBy{{.upperFields}}(ctx context.Context, {{.fieldAndDataTypes}}) error
{{- end }}