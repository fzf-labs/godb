// FindMultiCacheBy{{.upperFields}} 根据{{.upperFields}}查询多条数据，并设置缓存
func ({{.firstTableChar}} *{{.upperTableName}}Repo) FindMultiCacheBy{{.upperFields}}(ctx context.Context, {{.fieldAndDataTypes}}) ([]*{{.dbName}}_model.{{.upperTableName}}, error) {
	resp := make([]*{{.dbName}}_model.{{.upperTableName}}, 0)
	cacheKey := {{.firstTableChar}}.cache.Key(Cache{{.upperTableName}}By{{.cacheFields}}Prefix, {{.cacheFieldsJoin}})
	cacheValue, err := {{.firstTableChar}}.cache.Fetch(ctx, cacheKey, func() (string, error) {
		dao := {{.dbName}}_dao.Use({{.firstTableChar}}.db).{{.upperTableName}}
	    result, err := dao.WithContext(ctx).Where({{.whereFields}}).Find()
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return "", err
		}
        marshal, err := {{.firstTableChar}}.encoding.Marshal(result)
        if err != nil {
            return "", err
        }
		return string(marshal), nil
	}, {{.firstTableChar}}.cache.TTL())
	if err != nil {
		return nil, err
	}
	if cacheValue != "" {
		err = {{.firstTableChar}}.encoding.Unmarshal([]byte(cacheValue), &resp)
		if err != nil {
			return nil, err
		}
	}
	return resp, nil
}