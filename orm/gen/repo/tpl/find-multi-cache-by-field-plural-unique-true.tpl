// FindMultiCacheBy{{.upperFieldPlural}} 根据{{.lowerFieldPlural}}查询多条数据，并设置缓存
func ({{.firstTableChar}} *{{.upperTableName}}Repo) FindMultiCacheBy{{.upperFieldPlural}}(ctx context.Context, {{.lowerFieldPlural}} []{{.dataType}}) ([]*{{.dbName}}_model.{{.upperTableName}}, error) {
	resp := make([]*{{.dbName}}_model.{{.upperTableName}}, 0)
	cacheKeys := make([]string, 0)
	keyToParam := make(map[string]{{.dataType}})
	for _, v := range {{.lowerFieldPlural}} {
	    cacheKey := {{.firstTableChar}}.cache.Key(Cache{{.upperTableName}}By{{.upperField}}Prefix, v)
		cacheKeys = append(cacheKeys,cacheKey)
		keyToParam[cacheKey] = v
	}
	cacheValue, err := {{.firstTableChar}}.cache.FetchBatch(ctx, cacheKeys, func(miss []string) (map[string]string, error) {
        dbValue := make(map[string]string)
        params := make([]{{.dataType}},0)
        for _, v := range miss {
        	dbValue[v] = ""
            params = append(params, keyToParam[v])
        }
		dao := {{.dbName}}_dao.Use({{.firstTableChar}}.db).{{.upperTableName}}
		result, err := dao.WithContext(ctx).Where(dao.{{.upperField}}.In(params...)).Find()
		if err != nil {
			return nil, err
		}
		for _, v := range result {
            marshal, err := {{.firstTableChar}}.encoding.Marshal(v)
            if err != nil {
                return nil, err
            }
			dbValue[{{.firstTableChar}}.cache.Key( Cache{{.upperTableName}}By{{.upperField}}Prefix, v.{{.upperField}})] = string(marshal)
		}
		return dbValue, nil
	}, {{.firstTableChar}}.cache.TTL())
	if err != nil {
		return nil, err
	}
	for _, cacheKey := range cacheKeys {
		if cacheValue[cacheKey] != ""{
			tmp := new({{.dbName}}_model.{{.upperTableName}})
			err := {{.firstTableChar}}.encoding.Unmarshal([]byte(cacheValue[cacheKey]), tmp)
			if err != nil {
				return nil, err
			}
			resp = append(resp, tmp)
		}
	}
	return resp, nil
}