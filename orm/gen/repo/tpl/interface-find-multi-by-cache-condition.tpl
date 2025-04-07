// FindMultiCacheByCondition 自定义查询数据(通用),并设置缓存
FindMultiCacheByCondition(ctx context.Context, conditionReq *condition.Req) ([]*{{.dbName}}_model.{{.upperTableName}}, *condition.Reply, error)