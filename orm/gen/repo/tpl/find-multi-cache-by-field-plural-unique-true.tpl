// FindMultiCacheBy{{.upperFieldPlural}} 根据{{.lowerFieldPlural}}查询多条数据，并设置缓存
func ({{.firstTableChar}} *{{.upperTableName}}Repo) FindMultiCacheBy{{.upperFieldPlural}}(ctx context.Context, {{.lowerFieldPlural}} []{{.dataType}}) ([]*{{.dbName}}_model.{{.upperTableName}}, error) {
	resp := make([]*{{.dbName}}_model.{{.upperTableName}}, 0)
	cacheKeys := make([]string, 0)
	keyToParam := make(map[string]{{.dataType}})
	for _, item := range {{.lowerFieldPlural}} {
	    cacheKey := {{.firstTableChar}}.cache.Key(Cache{{.upperTableName}}By{{.upperField}}Prefix, item)
		cacheKeys = append(cacheKeys,cacheKey)
		keyToParam[cacheKey] = item
	}
	cacheValue, err := {{.firstTableChar}}.cache.FetchBatch(ctx, cacheKeys, func(miss []string) (map[string]string, error) {
        dbValue := make(map[string]string)
        params := make([]{{.dataType}},0)
        for _, item := range miss {
        	dbValue[item] = ""
            params = append(params, keyToParam[item])
        }
		dao := {{.dbName}}_dao.Use({{.firstTableChar}}.db).{{.upperTableName}}
		result, err := dao.WithContext(ctx).Where(dao.{{.upperField}}.In(params...)).Find()
		if err != nil {
			return nil, err
		}
		for _, item := range result {
            marshal, err := {{.firstTableChar}}.encoding.Marshal(item)
            if err != nil {
                return nil, err
            }
			dbValue[{{.firstTableChar}}.cache.Key( Cache{{.upperTableName}}By{{.upperField}}Prefix, item.{{.upperField}})] = string(marshal)
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
{{- if .haveDeletedAt }}
// FindMultiUnscopedCacheBy{{.upperFieldPlural}} 根据{{.lowerFieldPlural}}查询多条数据（包括软删除），并设置缓存
func ({{.firstTableChar}} *{{.upperTableName}}Repo) FindMultiUnscopedCacheBy{{.upperFieldPlural}}(ctx context.Context, {{.lowerFieldPlural}} []{{.dataType}}) ([]*{{.dbName}}_model.{{.upperTableName}}, error) {
	resp := make([]*{{.dbName}}_model.{{.upperTableName}}, 0)
	cacheKeys := make([]string, 0)
	keyToParam := make(map[string]{{.dataType}})
	for _, item := range {{.lowerFieldPlural}} {
	    cacheKey := {{.firstTableChar}}.cache.Key(Cache{{.upperTableName}}UnscopedBy{{.upperField}}Prefix, item)
		cacheKeys = append(cacheKeys,cacheKey)
		keyToParam[cacheKey] = item
	}
	cacheValue, err := {{.firstTableChar}}.cache.FetchBatch(ctx, cacheKeys, func(miss []string) (map[string]string, error) {
        dbValue := make(map[string]string)
        params := make([]{{.dataType}},0)
        for _, item := range miss {
        	dbValue[item] = ""
            params = append(params, keyToParam[item])
        }
		dao := {{.dbName}}_dao.Use({{.firstTableChar}}.db).{{.upperTableName}}
		result, err := dao.WithContext(ctx).Unscoped().Where(dao.{{.upperField}}.In(params...)).Find()
		if err != nil {
			return nil, err
		}
		for _, item := range result {
            marshal, err := {{.firstTableChar}}.encoding.Marshal(item)
            if err != nil {
                return nil, err
            }
			dbValue[{{.firstTableChar}}.cache.Key( Cache{{.upperTableName}}UnscopedBy{{.upperField}}Prefix, item.{{.upperField}})] = string(marshal)
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
{{- end }}