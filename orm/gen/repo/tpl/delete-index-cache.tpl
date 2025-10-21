// DeleteUniqueIndexCache 删除索引存在的缓存
func ({{.firstTableChar}} *{{.upperTableName}}Repo) DeleteIndexCache(ctx context.Context, data ...*{{.dbName}}_model.{{.upperTableName}}) error {
	KeyMap := make(map[string]struct{})
	keys := make([]string,0)
	keys = append(keys, {{.firstTableChar}}.cache.Key(Cache{{.upperTableName}}ByConditionPrefix))
	{{- if .haveDeletedAt }}
	keys = append(keys, {{.firstTableChar}}.cache.Key(Cache{{.upperTableName}}UnscopedByConditionPrefix))
	{{- end }}
	for _, item := range data {
		if item != nil {
			{{.cacheDelKeys}}
		}
	}
	for item := range KeyMap {
		keys = append(keys, item)
	}
	err := {{.firstTableChar}}.cache.DelBatch(ctx, keys)
	if err != nil {
		return err
	}
	return nil
}