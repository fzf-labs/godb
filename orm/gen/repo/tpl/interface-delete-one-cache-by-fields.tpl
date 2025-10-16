// DeleteOneCacheBy{{.upperFields}} 根据{{.upperFields}}删除一条数据，并删除缓存
DeleteOneCacheBy{{.upperFields}}(ctx context.Context, {{.fieldAndDataTypes}}) error
{{- if .haveDeletedAt }}
// DeleteOneUnscopedCacheBy{{.upperFields}} 根据{{.upperFields}}删除一条数据，并删除缓存
DeleteOneUnscopedCacheBy{{.upperFields}}(ctx context.Context, {{.fieldAndDataTypes}}) error
{{- end }}