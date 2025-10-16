// DeleteMultiCacheBy{{.upperFields}} 根据{{.lowerField}}删除多条数据，并删除缓存
DeleteMultiCacheBy{{.upperFields}}(ctx context.Context, {{.fieldAndDataTypes}}) error
{{- if .haveDeletedAt }}
// DeleteMultiUnscopedCacheBy{{.upperFields}} 根据{{.lowerField}}删除多条数据，并删除缓存
DeleteMultiUnscopedCacheBy{{.upperFields}}(ctx context.Context, {{.fieldAndDataTypes}}) error
{{- end }}