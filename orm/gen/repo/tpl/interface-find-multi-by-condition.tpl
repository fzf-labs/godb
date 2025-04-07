// FindMultiByCondition 自定义查询数据(通用)
FindMultiByCondition(ctx context.Context, conditionReq *condition.Req) ([]*{{.dbName}}_model.{{.upperTableName}}, *condition.Reply, error)