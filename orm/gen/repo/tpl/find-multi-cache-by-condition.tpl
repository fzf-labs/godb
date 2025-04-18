// FindMultiCacheByCondition 自定义查询数据(通用),并设置缓存
func ({{.firstTableChar}} *{{.upperTableName}}Repo) FindMultiCacheByCondition(ctx context.Context, conditionReq *condition.Req) ([]*{{.dbName}}_model.{{.upperTableName}}, *condition.Reply, error) {
	type Tmp struct {
		Result []*{{.dbName}}_model.{{.upperTableName}}
		ConditionReply *condition.Reply
	}
	tmp := Tmp{
		Result: make([]*{{.dbName}}_model.{{.upperTableName}}, 0),
		ConditionReply: &condition.Reply{},
	}
	cacheKey := {{.firstTableChar}}.cache.Key(Cache{{.upperTableName}}ByConditionPrefix)
	cacheField := conditionReq.ConvertToCacheField()
	cacheValue, err := {{.firstTableChar}}.cache.FetchHash(ctx, cacheKey,cacheField, func() (string, error) {
		result, conditionReply, err := {{.firstTableChar}}.FindMultiByCondition(ctx, conditionReq)
		if err != nil {
			return "", err
		}
		tmp.Result = result
		tmp.ConditionReply = conditionReply
		marshal, err := {{.firstTableChar}}.encoding.Marshal(tmp)
		if err != nil {
			return "", err
		}
		return string(marshal), nil
	}, {{.firstTableChar}}.cache.TTL())
	if err != nil {
		return tmp.Result, tmp.ConditionReply, err
	}
	if cacheValue != "" {
		err = {{.firstTableChar}}.encoding.Unmarshal([]byte(cacheValue), &tmp)
		if err != nil {
			return tmp.Result, tmp.ConditionReply, err
		}
	}
	return tmp.Result, tmp.ConditionReply, nil
}
