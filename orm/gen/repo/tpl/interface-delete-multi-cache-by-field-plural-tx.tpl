// DeleteMultiCacheBy{{.upperFieldPlural}}Tx 根据{{.upperFieldPlural}}删除多条数据，并删除缓存(事务)
DeleteMultiCacheBy{{.upperFieldPlural}}Tx(ctx context.Context,tx *{{.dbName}}_dao.Query, {{.lowerFieldPlural}} []{{.dataType}}) error
{{- if .haveDeletedAt }}
// DeleteMultiUnscopedCacheBy{{.upperFieldPlural}}Tx 根据{{.upperFieldPlural}}删除多条数据，并删除缓存(事务)
DeleteMultiUnscopedCacheBy{{.upperFieldPlural}}Tx(ctx context.Context,tx *{{.dbName}}_dao.Query, {{.lowerFieldPlural}} []{{.dataType}}) error
{{- end }}