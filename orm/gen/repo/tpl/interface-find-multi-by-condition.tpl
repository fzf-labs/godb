// FindMultiByCondition 自定义查询数据(通用)
FindMultiByCondition(ctx context.Context, conditionReq *condition.Req) ([]*{{.dbName}}_model.{{.upperTableName}}, *condition.Reply, error)
{{- if .haveDeletedAt }}
// FindMultiUnscopedByCondition 自定义查询数据(通用)（包括软删除）
FindMultiUnscopedByCondition(ctx context.Context, conditionReq *condition.Req) ([]*{{.dbName}}_model.{{.upperTableName}}, *condition.Reply, error)
{{- end }}