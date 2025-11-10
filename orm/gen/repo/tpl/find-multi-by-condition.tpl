// FindMultiByCondition 自定义查询数据(通用)
// 非万能查询方法,请评估后谨慎使用
func ({{.firstTableChar}} *{{.upperTableName}}Repo) FindMultiByCondition(ctx context.Context, conditionReq *condition.Req) ([]*{{.dbName}}_model.{{.upperTableName}}, *condition.Reply, error) {
	result := make([]*{{.dbName}}_model.{{.upperTableName}}, 0)
	conditionReply := &condition.Reply{}
	var total int64
	whereExpressions, orderExpressions, err := conditionReq.ConvertToGormExpression({{.dbName}}_model.{{.upperTableName}}{})
	if err != nil {
		return result, conditionReply, err
	}
	if conditionReq.Page != 0 && conditionReq.PageSize != 0 {
		err = {{.firstTableChar}}.db.WithContext(ctx).Model(&{{.dbName}}_model.{{.upperTableName}}{}).Clauses(whereExpressions...).Count(&total).Error
		if err != nil {
			return result, conditionReply, err
		}
		if total == 0 {
			return result, conditionReply, nil
		}
		conditionReply,err = conditionReq.ConvertToPage(int32(total))
		if err != nil {
			return result, conditionReply, err
		}
		query := {{.firstTableChar}}.db.WithContext(ctx).Model(&{{.dbName}}_model.{{.upperTableName}}{}).Clauses(whereExpressions...).Clauses(orderExpressions...)
		if conditionReply.Page != 0 && conditionReply.PageSize != 0 {
			query = query.Offset(int((conditionReply.Page - 1) * conditionReply.PageSize))
			query = query.Limit(int(conditionReply.PageSize))
		}
		err = query.Find(&result).Error
		if err != nil {
			return result, conditionReply, err
		}
	} else {
		err = {{.firstTableChar}}.db.WithContext(ctx).Model(&{{.dbName}}_model.{{.upperTableName}}{}).Clauses(whereExpressions...).Clauses(orderExpressions...).Find(&result).Error
		if err != nil {
			return result, conditionReply, err
		}
		conditionReply.Total = int32(len(result))
	}
	return result, conditionReply, err
}
{{- if .haveDeletedAt }}
// FindMultiUnscopedByCondition 自定义查询数据(通用)（包括软删除）
// 非万能查询方法,请评估后谨慎使用
func ({{.firstTableChar}} *{{.upperTableName}}Repo) FindMultiUnscopedByCondition(ctx context.Context, conditionReq *condition.Req) ([]*{{.dbName}}_model.{{.upperTableName}}, *condition.Reply, error) {
	result := make([]*{{.dbName}}_model.{{.upperTableName}}, 0)
	conditionReply := &condition.Reply{}
	var total int64
	whereExpressions, orderExpressions, err := conditionReq.ConvertToGormExpression({{.dbName}}_model.{{.upperTableName}}{})
	if err != nil {
		return result, conditionReply, err
	}
	if conditionReq.Page != 0 && conditionReq.PageSize != 0 {
		err = {{.firstTableChar}}.db.WithContext(ctx).Model(&{{.dbName}}_model.{{.upperTableName}}{}).Unscoped().Clauses(whereExpressions...).Count(&total).Error
		if err != nil {
			return result, conditionReply, err
		}
		if total == 0 {
			return result, conditionReply, nil
		}
		conditionReply, err = conditionReq.ConvertToPage(int32(total))
		if err != nil {
			return result, conditionReply, err
		}
		query := {{.firstTableChar}}.db.WithContext(ctx).Model(&{{.dbName}}_model.{{.upperTableName}}{}).Unscoped().Clauses(whereExpressions...).Clauses(orderExpressions...)
		if conditionReply.Page != 0 && conditionReply.PageSize != 0 {
			query = query.Offset(int((conditionReply.Page - 1) * conditionReply.PageSize))
			query = query.Limit(int(conditionReply.PageSize))
		}
		err = query.Find(&result).Error
		if err != nil {
			return result, conditionReply, err
		}
	} else {
		err = {{.firstTableChar}}.db.WithContext(ctx).Model(&{{.dbName}}_model.{{.upperTableName}}{}).Unscoped().Clauses(whereExpressions...).Clauses(orderExpressions...).Find(&result).Error
		if err != nil {
			return result, conditionReply, err
		}
		conditionReply.Total = int32(len(result))
	}
	return result, conditionReply, err
}
{{- end }}
