// DeleteMultiCacheBy{{.upperField}} 根据{{.lowerField}}删除多条数据，并删除缓存
DeleteMultiCacheBy{{.upperField}}(ctx context.Context, {{.lowerField}} {{.dataType}}) error
{{- if .haveDeletedAt }}
// DeleteMultiUnscopedCacheBy{{.upperField}} 根据{{.lowerField}}删除多条数据，并删除缓存
DeleteMultiUnscopedCacheBy{{.upperField}}(ctx context.Context, {{.lowerField}} {{.dataType}}) error
{{- end }}