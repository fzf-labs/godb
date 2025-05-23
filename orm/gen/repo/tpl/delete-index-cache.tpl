// DeleteUniqueIndexCache 删除索引存在的缓存
func ({{.firstTableChar}} *{{.upperTableName}}Repo) DeleteIndexCache(ctx context.Context, data ...*{{.dbName}}_model.{{.upperTableName}}) error {
	KeyMap := make(map[string]struct{})
	keys := make([]string,0)
	keys = append(keys, {{.firstTableChar}}.cache.Key(Cache{{.upperTableName}}ByConditionPrefix))
	for _, v := range data {
		if v != nil {
			{{.cacheDelKeys}}
		}
	}
	for k := range KeyMap {
		keys = append(keys, k)
	}
	err := {{.firstTableChar}}.cache.DelBatch(ctx, keys)
	if err != nil {
		return err
	}
	return nil
}