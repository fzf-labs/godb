// DeleteMultiCacheBy{{.upperFieldPlural}} 根据{{.upperFieldPlural}}删除多条数据，并删除缓存
DeleteMultiCacheBy{{.upperFieldPlural}}(ctx context.Context, {{.lowerFieldPlural}} []{{.dataType}}) error