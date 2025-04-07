// UpdateBatchBy{{.upperSinglePrimaryKeyPlural}} 根据主键{{.upperSinglePrimaryKeyPlural}}批量更新
// 零值会被更新
func ({{.firstTableChar}} *{{.upperTableName}}Repo) UpdateBatchBy{{.upperSinglePrimaryKeyPlural}}(ctx context.Context, {{.lowerSinglePrimaryKeyPlural}} []{{.dataTypeSinglePrimaryKey}}, data map[string]interface{}) error {
	dao := {{.dbName}}_dao.Use({{.firstTableChar}}.db).{{.upperTableName}}
	_, err := dao.WithContext(ctx).Where(dao.{{.upperSinglePrimaryKey}}.In({{.lowerSinglePrimaryKeyPlural}}...)).Updates(data)
	if err != nil {
		return err
	}
	return nil
}