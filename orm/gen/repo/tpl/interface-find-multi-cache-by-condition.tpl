// FindMultiCacheByCondition 自定义查询数据(通用),并设置缓存
FindMultiCacheByCondition(ctx context.Context, conditionReq *condition.Req) ([]*{{.dbName}}_model.{{.upperTableName}}, *condition.Reply, error)
{{- if .haveDeletedAt }}
// FindMultiUnscopedCacheByCondition 自定义查询数据(通用)（包括软删除）,并设置缓存
FindMultiUnscopedCacheByCondition(ctx context.Context, conditionReq *condition.Req) ([]*{{.dbName}}_model.{{.upperTableName}}, *condition.Reply, error)
{{- end }}