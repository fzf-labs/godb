// UpdateBatchBy{{.upperFieldPlural}}Tx 根据字段{{.upperFieldPlural}}批量更新(事务),零值会被更新
func ({{.firstTableChar}} *{{.upperTableName}}Repo) UpdateBatchBy{{.upperFieldPlural}}Tx(ctx context.Context,tx *{{.dbName}}_dao.Query, {{.lowerFieldPlural}} []{{.dataType}}, data map[string]interface{}) error {
	dao := tx.{{.upperTableName}}
	_, err := dao.WithContext(ctx).Where(dao.{{.upperField}}.In({{.lowerFieldPlural}}...)).Updates(data)
	if err != nil {
		return err
	}
	return nil
}