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
		keyToValues := make(map[string][]*{{.dbName}}_model.{{.upperTableName}})
		for _, v := range result {
			key := {{.firstTableChar}}.cache.Key(Cache{{.upperTableName}}By{{.upperField}}Prefix, v.{{.upperField}})
			if keyToValues[key] == nil {
				keyToValues[key] = make([]*{{.dbName}}_model.{{.upperTableName}}, 0)
			}
			keyToValues[key] = append(keyToValues[key], v)
		}
		for k := range dbValue {
			if keyToValues[k] != nil {
				marshal, err := {{.firstTableChar}}.encoding.Marshal(keyToValues[k])
				if err != nil {
					return nil, err
				}
				dbValue[k] = string(marshal)
			}
		}
		return dbValue, nil
	}, {{.firstTableChar}}.cache.TTL())
	if err != nil {
		return nil, err
	}
	for _, v := range {{.lowerFieldPlural}} {
	    cacheKey := {{.firstTableChar}}.cache.Key(Cache{{.upperTableName}}By{{.upperField}}Prefix, v)
		if cacheValue[cacheKey] != ""{
			tmp := make([]*{{.dbName}}_model.{{.upperTableName}}, 0)
			err := {{.firstTableChar}}.encoding.Unmarshal([]byte(cacheValue[cacheKey]), &tmp)
			if err != nil {
				return nil, err
			}
			resp = append(resp, tmp...)
		}
	}
	return resp, nil
}