// FindOneCacheBy{{.upperField}} 根据{{.lowerField}}查询一条数据，并设置缓存
func ({{.firstTableChar}} *{{.upperTableName}}Repo) FindOneCacheBy{{.upperField}}(ctx context.Context, {{.lowerField}} {{.dataType}}) (*{{.dbName}}_model.{{.upperTableName}}, error) {
	resp := new({{.dbName}}_model.{{.upperTableName}})
	cacheKey := {{.firstTableChar}}.cache.Key(Cache{{.upperTableName}}By{{.cacheFields}}Prefix,{{.cacheFieldsJoin}})
	cacheValue, err := {{.firstTableChar}}.cache.Fetch(ctx, cacheKey, func() (string, error) {
		dao := {{.dbName}}_dao.Use({{.firstTableChar}}.db).{{.upperTableName}}
		result, err := dao.WithContext(ctx).Where({{.whereFields}}).First()
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
		err = {{.firstTableChar}}.encoding.Unmarshal([]byte(cacheValue), resp)
		if err != nil {
			return nil, err
		}
	}
	return resp, nil
}
{{- if .haveDeletedAt }}
// FindOneUnscopedCacheBy{{.upperField}} 根据{{.lowerField}}查询一条数据（包括软删除），并设置缓存
func ({{.firstTableChar}} *{{.upperTableName}}Repo) FindOneUnscopedCacheBy{{.upperField}}(ctx context.Context, {{.lowerField}} {{.dataType}}) (*{{.dbName}}_model.{{.upperTableName}}, error) {
	resp := new({{.dbName}}_model.{{.upperTableName}})
	cacheKey := {{.firstTableChar}}.cache.Key(Cache{{.upperTableName}}UnscopedBy{{.cacheFields}}Prefix,{{.cacheFieldsJoin}})
	cacheValue, err := {{.firstTableChar}}.cache.Fetch(ctx, cacheKey, func() (string, error) {
		dao := {{.dbName}}_dao.Use({{.firstTableChar}}.db).{{.upperTableName}}
		result, err := dao.WithContext(ctx).Unscoped().Where({{.whereFields}}).First()
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
		err = {{.firstTableChar}}.encoding.Unmarshal([]byte(cacheValue), resp)
		if err != nil {
			return nil, err
		}
	}
	return resp, nil
}
{{- end }}